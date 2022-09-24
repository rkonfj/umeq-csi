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
	return a.kv.Set(volumeId, []byte(kind))
}

func (a *Agent) removeVolumeKind(volumeId string) error {
	return a.kv.Del(volumeId)
}

func (a *Agent) lookupVolumePath(volumeId string) (string, error) {
	b, err := a.kv.Get(volumeId)
	if err != nil {
		return "", err
	}
	return a.storage[string(b)] + volumeId + ".qcow2", nil
}

func (a *Agent) UnpublishVolume(volumeId, nodeId string) error {
	return a.attacher.Detach(nodeId, volumeId)
}

func (a *Agent) PublishVolume(volumeId, nodeId string) error {
	qcow2Path, err := a.lookupVolumePath(volumeId)
	if err != nil {
		return err
	}
	return a.attacher.Attach(nodeId, volumeId, qcow2Path)
}

func (a *Agent) CreateVolume(kind, volumeId string, requiredBytes int64) error {
	qcowPath := a.storage[kind] + volumeId + ".qcow2"
	err := a.saveVolumeKind(volumeId, kind)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(qcowPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("volume %s alredy exists", volumeId)
	}
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create qcow2 err: %w %s", err, out)
	} else {
		log.Println("create qcow2:", string(out))
	}
	return nil
}

func (a *Agent) ExpandVolume(volumeId string, requiredBytes int64) error {
	qcowPath, err := a.lookupVolumePath(volumeId)
	if err != nil {
		return err
	}
	cmd := exec.Command("qemu-img", "resize", qcowPath, fmt.Sprintf("%d", requiredBytes))
	if out, err := cmd.Output(); err != nil {
		return err
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func (a *Agent) DeleteVolume(volumeId string) error {
	qcowPath, err := a.lookupVolumePath(volumeId)
	if err != nil {
		return err
	}
	if err := os.Remove(qcowPath); err != nil {
		return fmt.Errorf("delete qcow2 err:%w", err)
	}
	if err := a.attacher.Clean(volumeId); err != nil {
		log.Println("[warn] attacher clean failed:" + err.Error())
	}
	if err := a.removeVolumeKind(volumeId); err != nil {
		log.Println("[warn] volume kind state remove failed")
	}
	log.Println("Removed volume:", volumeId)
	return nil
}

func (a *Agent) GetDevPath(volumeId string) (string, error) {
	return a.attacher.DevPath(volumeId)
}
