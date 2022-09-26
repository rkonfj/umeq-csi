package attach

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tasselsd/umeq-csi/internel/qmp"
	"github.com/tasselsd/umeq-csi/internel/state"
)

type QmpAttacher struct {
	CommonAttacher
	mons map[string]*qmp.Monitor
}

type Sock struct {
	Name string
	Path string
}

func NewQmpAttacher(kv state.KvStore, qs []Sock) *QmpAttacher {
	attacher := &QmpAttacher{
		CommonAttacher: CommonAttacher{
			kv: kv,
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
	if len(attacher.mons) == 0 {
		panic("no any normal mons found, exiting...")
	}
	return attacher
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
	log.Printf("[qmp] node: %s, cmd: %s, out: %s\n", node, cmd, out)
	return nil
}

func (q *QmpAttacher) Attach(nodeId, volumeId, qcow2Path string) error {
	q.kv.Lock(volumeId)
	defer q.kv.Unlock(volumeId)
	log.Println("[info] qmp request attach", nodeId, volumeId, qcow2Path)
	cmd := fmt.Sprintf("drive_add 0 if=none,format=qcow2,file=%s,id=%s", qcow2Path, volumeId)
	err := q.exec(nodeId, cmd)
	if err != nil {
		log.Println("[error] exiting", err)
		os.Exit(1)
	}

	serialId, err := q.getSerialId(volumeId)
	if err != nil {
		return err
	}
	cmd2 := fmt.Sprintf("device_add virtio-blk-pci,drive=%s,id=%s,serial=%s",
		volumeId, volumeId, serialId)
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
