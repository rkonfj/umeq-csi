package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Agent struct {
	etcdcli  *clientv3.Client
	diskRoot string
	attacher Attacher
}

func NewAgent(etcdcli *clientv3.Client, diskRoot string, attacher Attacher) *Agent {
	return &Agent{
		etcdcli:  etcdcli,
		diskRoot: diskRoot,
		attacher: attacher,
	}
}

func (a *Agent) UnpublishVolume(volumeId, nodeId string) error {
	err := a.attacher.Exec(nodeId, "device_del "+volumeId)
	if err != nil {
		err = a.attacher.Exec(nodeId, "drive_del "+volumeId)
		if err != nil {
			return fmt.Errorf("unpushlish err:%w", err)
		}
	}
	return nil
}

func (a *Agent) PublishVolume(volumeId, nodeId string) error {
	qcow2Path := a.diskRoot + volumeId + ".qcow2"
	err := a.attacher.Exec(nodeId, fmt.Sprintf("drive_add 0 if=none,format=qcow2,file=%s,id=%s", qcow2Path, volumeId))
	if err != nil {
		return fmt.Errorf("publish[drive_add] err:%w", err)
	}

	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := a.etcdcli.Get(c, "/xiaomakai/"+volumeId)
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		id := a.nextID()
		a.etcdcli.Put(c, "/xiaomakai/"+volumeId, id)
		r, err = a.etcdcli.Get(c, "/xiaomakai/"+volumeId)
		if err != nil {
			panic(err)
		}
	}

	err = a.attacher.Exec(nodeId, fmt.Sprintf("device_add virtio-blk-pci,drive=%s,id=%s,serial=%s", volumeId, volumeId, r.Kvs[0].Value))
	if err != nil {
		err = a.attacher.Exec(nodeId, "drive_del "+volumeId)
		if err != nil {
			log.Println("rollback error:", err.Error())
		}
		return fmt.Errorf("publish[device_add] err:%w", err)
	}
	return nil
}

func (a *Agent) CreateVolume(volumeId string, requiredBytes int64) error {
	qcowPath := a.diskRoot + volumeId + ".qcow2"
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf("create qcow2 err:%w", err)
	} else {
		log.Println("create qcow2:", string(out))
	}
	return nil
}

func (a *Agent) ExpandVolume(volumeId string, requiredBytes int64) error {
	qcowPath := a.diskRoot + volumeId + ".qcow2"
	cmd := exec.Command("qemu-img", "resize", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.Output(); err != nil {
		return err
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func (a *Agent) DeleteVolume(volumeId string) error {
	err := os.Remove(a.diskRoot + volumeId + ".qcow2")
	if err != nil {
		return fmt.Errorf("delete qcow2 err:%w", err)
	}
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := a.etcdcli.Delete(c, "/xiaomakai/"+volumeId)
	if err != nil {
		log.Println("etcd delete ERR:", err)
	} else {
		log.Printf("etcd resp:%v\n", resp)
	}
	log.Println("Removed volume:", volumeId)
	return nil
}

func (a *Agent) GetDevPath(volumeId string) (string, error) {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := a.etcdcli.Get(c, "/xiaomakai/"+volumeId)
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		return "", fmt.Errorf("volume %s not found! not published yet?", volumeId)
	}
	return "/dev/disk/by-id/virtio-" + string(r.Kvs[0].Value), nil
}

var m sync.Mutex

func (a *Agent) nextID() string {
	m.Lock()
	defer m.Unlock()
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := a.etcdcli.Get(c, "/xiaomakai/id")
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		a.etcdcli.Put(c, "/xiaomakai/id", "1")
	} else {
		value, _ := strconv.Atoi(string(r.Kvs[0].Value))
		value += 1
		a.etcdcli.Put(c, "/xiaomakai/id", fmt.Sprintf("%d", value))
		return string(r.Kvs[0].Value)
	}
	return "0"
}
