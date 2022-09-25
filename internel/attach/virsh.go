package attach

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/tasselsd/umeq-csi/internel/state"
)

type VirshAttacher struct {
	CommonAttacher
}

func NewVirshAttacher(kv state.KvStore) *VirshAttacher {
	return &VirshAttacher{
		CommonAttacher: CommonAttacher{
			kv: kv,
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
		if len(target) == 0 || !strings.HasPrefix(target, "vd") {
			continue
		}
		if target[2] == targetLetter {
			targetLetter = targetLetter + 1
			goto probe
		}
	}
	target := "vd" + string(targetLetter)
	log.Println("determined target:", target)
	return target, nil
}

func (v *VirshAttacher) lookupTarget(volumeId string) (string, error) {
	resp, err := v.kv.Get("/xiaomakai/" + volumeId)
	if err != nil {
		return "", fmt.Errorf("[error] virsh not attach %s yet?", volumeId)
	}
	return string(resp), nil
}

func (v *VirshAttacher) Attach(nodeId, volumeId, qcow2Path string) error {
	log.Println("[info] virsh request attach", nodeId, volumeId, qcow2Path)
	v.kv.Lock(nodeId)
	defer v.kv.Unlock(nodeId)
	var target string
	r, err := v.kv.Get("/xiaomakai/" + volumeId)
	if err != nil {
		_target, err := v.target(nodeId)
		if err != nil {
			return err
		}
		err = v.kv.Set("/xiaomakai/"+volumeId, []byte(_target))
		if err != nil {
			return fmt.Errorf("[error] kvStore set err: %w", err)
		}
		target = _target
	} else {
		target = string(r)
	}

	cmd := fmt.Sprintf("virsh attach-disk %s %s %s --driver qemu --subdriver qcow2 --targetbus virtio",
		nodeId, qcow2Path, target)
	log.Println(cmd)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("[virsh] attach-disk failed %s error:%w", string(out), err)
	}
	log.Println(string(out))
	return nil
}

func (v *VirshAttacher) Detach(nodeId, volumeId string) error {
	log.Println("[info] virsh request detach", nodeId, volumeId)
	target, err := v.lookupTarget(volumeId)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("virsh detach-disk %s %s", nodeId, target)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("[virsh] detach-disk failed %s error:%w", out, err)
	}
	log.Println(string(out))
	return nil
}

func (v *VirshAttacher) DevPath(volumeId string) (string, error) {
	target, err := v.lookupTarget(volumeId)
	if err != nil {
		return "", err
	}
	return "/dev/" + target, nil
}
