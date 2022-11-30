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
	"os"
	"time"

	"github.com/tasselsd/umeq-csi/pkg/qmp"
	"github.com/tasselsd/umeq-csi/pkg/state"
)

// Qmp protocol attacher
type QmpAttacher struct {
	CommonAttacher

	// qemu qmp monitor unix socket operators
	mons map[string]*qmp.Monitor
}

// Socket for qemu virtual machine
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

// execute human read qmp command on node
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
