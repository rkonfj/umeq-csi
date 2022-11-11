package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/tasselsd/umeq-csi/internel/attach"
	"github.com/tasselsd/umeq-csi/internel/state"
)

type Agent struct {
	storage  map[string]string
	kv       state.KvStore
	attacher attach.Attacher
}

func NewAgent(storage map[string]string, kv state.KvStore, attacher attach.Attacher) *Agent {
	return &Agent{
		storage:  storage,
		kv:       kv,
		attacher: attacher,
	}
}

func (a *Agent) saveVolumeKind(volumeId, kind string) error {
	return a.kv.Set("kind-"+volumeId, []byte(kind))
}

func (a *Agent) removeVolumeKind(volumeId string) error {
	return a.kv.Del("kind-" + volumeId)
}

func (a *Agent) lookupVolumePath(volumeId string) string {
	kind := "default"
	b, err := a.kv.Get("kind-" + volumeId)
	if err != nil {
		log.Println("[warn] lookupVolumePath failed, fallback to [default]:", err)
	} else {
		kind = string(b)
	}
	log.Println("[info] lookup volumeId", volumeId, "kind is", kind)
	return a.storage[kind] + volumeId + ".qcow2"
}

func (a *Agent) UnpublishVolume(volumeId, nodeId string) error {
	return a.attacher.Detach(nodeId, volumeId)
}

func (a *Agent) PublishVolume(volumeId, nodeId string) error {
	qcow2Path := a.lookupVolumePath(volumeId)
	return a.attacher.Attach(nodeId, volumeId, qcow2Path)
}

func (a *Agent) CreateVolume(kind, volumeId string, requiredBytes int64) error {
	if len(kind) == 0 {
		kind = "default"
	}
	if _, ok := a.storage[kind]; !ok {
		return fmt.Errorf("[error] volume %s create failed, storage kind %s not found",
			volumeId, kind)
	}
	qcowPath := a.storage[kind] + volumeId + ".qcow2"
	err := a.saveVolumeKind(volumeId, kind)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(qcowPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("volume %s alredy exists", volumeId)
	}
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", qcowPath,
		fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create qcow2 err: %w %s", err, out)
	} else {
		log.Println("create qcow2:", string(out))
	}
	return nil
}

func (a *Agent) ExpandVolume(volumeId string, requiredBytes int64) error {
	qcowPath := a.lookupVolumePath(volumeId)
	cmd := exec.Command("qemu-img", "resize", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.Output(); err != nil {
		return err
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func (a *Agent) DeleteVolume(volumeId string) error {
	qcowPath := a.lookupVolumePath(volumeId)
	if err := os.Remove(qcowPath); err != nil {
		return fmt.Errorf("delete qcow2 err:%w", err)
	}
	if err := a.removeVolumeKind(volumeId); err != nil {
		log.Println("[warn] volume kind state remove failed", err)
	}
	log.Println("Removed volume:", volumeId)
	return nil
}

func (a *Agent) GetDevPath(volumeId string) (string, error) {
	return a.attacher.DevPath(volumeId)
}
