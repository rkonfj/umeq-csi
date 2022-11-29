// Copyright 2022 rkonfj@fnla.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package attach

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/tasselsd/umeq-csi/pkg/state"
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
		return "", fmt.Errorf("[error] virsh domblklist error %w", err)
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
	log.Println("determined target[", target, "] on node ", nodeId)
	return target, nil
}

func (v *VirshAttacher) lookupTarget(volumeId string) (string, error) {
	resp, err := v.kv.Get("/xiaomakai/virsh/target/" + volumeId)
	if err != nil {
		return "", fmt.Errorf("[error] virsh not attach %s yet?", volumeId)
	}
	return string(resp), nil
}

func (v *VirshAttacher) generateTarget(nodeId, volumeId string) (string, error) {
	r, err := v.kv.Get("/xiaomakai/virsh/target/" + volumeId)
	if err != nil {
		_target, err := v.target(nodeId)
		if err != nil {
			return "", err
		}
		err = v.kv.Set("/xiaomakai/virsh/target/"+volumeId, []byte(_target))
		if err != nil {
			return "", fmt.Errorf("[error] kvStore set err: %w", err)
		}
		return _target, nil
	}
	return string(r), nil
}

func (v *VirshAttacher) Attach(nodeId, volumeId, qcow2Path string) error {
	log.Println("[info] virsh request attach", nodeId, volumeId, qcow2Path)
	v.kv.Lock(volumeId)
	defer v.kv.Unlock(volumeId)
	target, err := v.generateTarget(nodeId, volumeId)
	if err != nil {
		return err
	}
	serialId, err := v.getSerialId(volumeId)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf(
		"virsh attach-disk %s %s %s --driver qemu --subdriver qcow2 --targetbus virtio --serial %s",
		nodeId, qcow2Path, target, serialId)
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
	v.Clean(volumeId)
	return nil
}
func (v *VirshAttacher) Clean(volumeId string) error {
	err1 := v.CommonAttacher.Clean(volumeId)
	err2 := v.kv.Del("/xiaomakai/virsh/target/" + volumeId)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
