package attach

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tasselsd/umeq-csi/internel/qmp"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type QmpAttacher struct {
	CommonAttacher
	mons map[string]*qmp.Monitor
}

type Sock struct {
	Name string
	Path string
}

func NewQmpAttacher(etcdctl *clientv3.Client, qs []Sock) *QmpAttacher {
	attacher := &QmpAttacher{
		CommonAttacher: CommonAttacher{
			etcdctl: etcdctl,
		},
		mons: make(map[string]*qmp.Monitor),
	}

	for _, q := range qs {
		mon, err := qmp.NewMonitor(q.Path, 60*time.Second)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Registered %s -> %s\n", q.Name, q.Path)
		attacher.mons[q.Name] = mon
	}

	return attacher
}

func (q *QmpAttacher) nextSeq() string {
	q.lock("/global")
	defer q.unlock("/global")
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := q.etcdctl.Get(c, "/xiaomakai/id")
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		q.etcdctl.Put(c, "/xiaomakai/id", "1")
	} else {
		value, _ := strconv.Atoi(string(r.Kvs[0].Value))
		value += 1
		q.etcdctl.Put(c, "/xiaomakai/id", fmt.Sprintf("%d", value))
		return string(r.Kvs[0].Value)
	}
	return "0"
}

func (q *QmpAttacher) exec(node, cmd string) error {
	var out string

	mon := q.mons[node]

	if mon == nil {
		return fmt.Errorf("mon %s not found", node)
	}

	if err := mon.Run(qmp.Command{
		Name: "human-monitor-command",
		Arguments: &qmp.HumanCommand{
			Cmd: cmd,
		},
	}, &out); err != nil {
		return err
	}
	log.Printf("[qmp]node: %s, cmd: %s, out: %s\n", node, cmd, out)
	return nil
}

func (q *QmpAttacher) Attach(nodeId, volumeId, qcow2Path string) error {
	log.Println("[info] qmp request attach", nodeId, volumeId, qcow2Path)
	cmd := fmt.Sprintf("drive_add 0 if=none,format=qcow2,file=%s,id=%s", qcow2Path, volumeId)
	err := q.exec(nodeId, cmd)
	if err != nil {
		return fmt.Errorf("attach[drive_add] err:%w", err)
	}

	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := q.etcdctl.Get(c, "/xiaomakai/"+volumeId)
	if err != nil {
		return err
	}
	if r.Count == 0 {
		id := q.nextSeq()
		q.etcdctl.Put(c, "/xiaomakai/"+volumeId, id)
		r, err = q.etcdctl.Get(c, "/xiaomakai/"+volumeId)
		if err != nil {
			return err
		}
	}
	cmd2 := fmt.Sprintf("device_add virtio-blk-pci,drive=%s,id=%s,serial=%s",
		volumeId, volumeId, r.Kvs[0].Value)
	err = q.exec(nodeId, cmd2)
	if err != nil {
		err = q.exec(nodeId, "drive_del "+volumeId)
		if err != nil {
			log.Println("rollback error:", err.Error())
		}
		return fmt.Errorf("attach[device_add] err:%w", err)
	}
	return nil
}

func (q *QmpAttacher) Detach(nodeId, volumeId string) error {
	log.Println("[info] qmp request detach", nodeId, volumeId)
	err := q.exec(nodeId, "device_del "+volumeId)
	if err != nil {
		err = q.exec(nodeId, "drive_del "+volumeId)
		if err != nil {
			return fmt.Errorf("detach err:%w", err)
		}
	}
	q.Clean(volumeId)
	return nil
}

func (q *QmpAttacher) DevPath(volumeId string) (string, error) {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := q.etcdctl.Get(c, "/xiaomakai/"+volumeId)
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		return "", fmt.Errorf("volume %s not found! not attach yet?", volumeId)
	}
	return "/dev/disk/by-id/virtio-" + string(r.Kvs[0].Value), nil
}
