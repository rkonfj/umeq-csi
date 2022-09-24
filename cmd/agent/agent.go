package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/tasselsd/umeq-csi/internel/attach"
)

type Agent struct {
	diskRoot string
	attacher attach.Attacher
}

func NewAgent(diskRoot string, attacher attach.Attacher) *Agent {
	return &Agent{
		diskRoot: diskRoot,
		attacher: attacher,
	}
}

func (a *Agent) UnpublishVolume(volumeId, nodeId string) error {
	return a.attacher.Detach(nodeId, volumeId)
}

func (a *Agent) PublishVolume(volumeId, nodeId string) error {
	qcow2Path := a.diskRoot + volumeId + ".qcow2"
	return a.attacher.Attach(nodeId, volumeId, qcow2Path)
}

func (a *Agent) CreateVolume(volumeId string, requiredBytes int64) error {
	qcowPath := a.diskRoot + volumeId + ".qcow2"
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
	err = a.attacher.Clean(volumeId)
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}
	log.Println("Removed volume:", volumeId)
	return nil
}

func (a *Agent) GetDevPath(volumeId string) (string, error) {
	return a.attacher.DevPath(volumeId)
}
