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
	"go.etcd.io/etcd/pkg/transport"
)

var diskRoot string = "/fs/trust/vm/csi/"
var etcdcli *clientv3.Client

func init() {
	tlsInfo := transport.TLSInfo{
		CertFile:      "etcd.crt",
		KeyFile:       "etcd.key",
		TrustedCAFile: "etcd-ca.crt",
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.3.35:2379"},
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}
	etcdcli = cli
}

func DoUnpublishVolume(volumeId, nodeId string) error {
	err := Exec(nodeId, "device_del "+volumeId)
	if err != nil {
		err = Exec(nodeId, "drive_del "+volumeId)
		if err != nil {
			return fmt.Errorf("unpushlish err:%w", err)
		}
	}
	return nil
}

func DoPublishVolume(volumeId, nodeId string) error {
	qcow2Path := diskRoot + volumeId + ".qcow2"
	err := Exec(nodeId, fmt.Sprintf("drive_add 0 if=none,format=qcow2,file=%s,id=%s", qcow2Path, volumeId))
	if err != nil {
		return fmt.Errorf("publish[drive_add] err:%w", err)
	}

	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := etcdcli.Get(c, "/xiaomakai/"+volumeId)
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		id := nextID()
		etcdcli.Put(c, "/xiaomakai/"+volumeId, id)
		r, err = etcdcli.Get(c, "/xiaomakai/"+volumeId)
		if err != nil {
			panic(err)
		}
	}

	err = Exec(nodeId, fmt.Sprintf("device_add virtio-blk-pci,drive=%s,id=%s,serial=%s", volumeId, volumeId, r.Kvs[0].Value))
	if err != nil {
		err = Exec(nodeId, "drive_del "+volumeId)
		if err != nil {
			log.Println("rollback error:", err.Error())
		}
		return fmt.Errorf("publish[device_add] err:%w", err)
	}
	return nil
}

func DoCreateVolume(volumeId string, requiredBytes int64) error {
	qcowPath := diskRoot + volumeId + ".qcow2"
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf("create qcow2 err:%w", err)
	} else {
		log.Println("create qcow2:", string(out))
	}
	return nil
}

func DoExpandVolume(volumeId string, requiredBytes int64) error {
	qcowPath := diskRoot + volumeId + ".qcow2"
	cmd := exec.Command("qemu-img", "resize", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.Output(); err != nil {
		return err
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func DoDeleteVolume(volumeId string) error {
	err := os.Remove(diskRoot + volumeId + ".qcow2")
	if err != nil {
		return fmt.Errorf("delete qcow2 err:%w", err)
	}
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := etcdcli.Delete(c, "/xiaomakai/"+volumeId)
	if err != nil {
		log.Println("etcd delete ERR:", err)
	} else {
		log.Printf("etcd resp:%v\n", resp)
	}
	log.Println("Removed volume:", volumeId)
	return nil
}

func DoGetDevPath(volumeId string) (string, error) {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := etcdcli.Get(c, "/xiaomakai/"+volumeId)
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		return "", fmt.Errorf("volume %s not found! not published yet?", volumeId)
	}
	return "/dev/disk/by-id/virtio-" + string(r.Kvs[0].Value), nil
}

var m sync.Mutex

func nextID() string {
	m.Lock()
	defer m.Unlock()
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := etcdcli.Get(c, "/xiaomakai/id")
	if err != nil {
		panic(err)
	}
	if r.Count == 0 {
		etcdcli.Put(c, "/xiaomakai/id", "1")
	} else {
		value, _ := strconv.Atoi(string(r.Kvs[0].Value))
		value += 1
		etcdcli.Put(c, "/xiaomakai/id", fmt.Sprintf("%d", value))
		return string(r.Kvs[0].Value)
	}
	return "0"
}
