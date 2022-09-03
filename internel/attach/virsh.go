package attach

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type VirshAttacher struct {
	CommonAttacher
}

func NewVirshAttacher(etcdctl *clientv3.Client) *VirshAttacher {
	return &VirshAttacher{
		CommonAttacher: CommonAttacher{
			etcdctl: etcdctl,
		},
	}
}

func (v *VirshAttacher) target(nodeId string) (string, error) {
	cmd := fmt.Sprintf("virsh domblklist %s | tail -n +3 | awk '{print $1}'", nodeId)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	existsTargets := strings.Split(string(out), "\n")
	targetLetter := byte('a')
probe:
	for _, target := range existsTargets {
		if !strings.HasPrefix(target, "vd") {
			continue
		}
		if target[2] == byte(targetLetter) {
			targetLetter = targetLetter + 1
			break probe
		}
	}
	return "vd" + string(targetLetter+1), nil
}

func (v *VirshAttacher) targetFromEtcd(volumeId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := v.etcdctl.Get(ctx, "/xiaomakai/"+volumeId)
	if err != nil {
		return "", err
	}
	if resp.Count == 0 {
		return "", fmt.Errorf("not attach %s yet?", volumeId)
	}
	return string(resp.Kvs[0].Value), nil
}

func (v *VirshAttacher) Attach(nodeId, volumeId, qcow2Path string) error {
	v.lock(nodeId)
	defer v.unlock(nodeId)
	taregt, err := v.target(nodeId)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = v.etcdctl.Put(ctx, "/xiaomakai/"+volumeId, taregt)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("virsh attach-disk %s %s %s --driver qemu --subdriver qcow2 --targetbus virtio",
		nodeId, qcow2Path, taregt)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return err
	}
	log.Println(string(out))
	return nil
}

func (v *VirshAttacher) Detach(nodeId, volumeId string) error {
	target, err := v.targetFromEtcd(volumeId)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("virsh detach-disk %s %s", nodeId, target)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return err
	}
	log.Println(string(out))
	return nil
}

func (v *VirshAttacher) DevPath(volumeId string) (string, error) {
	target, err := v.targetFromEtcd(volumeId)
	if err != nil {
		return "", err
	}
	return "/dev/" + target, nil
}
