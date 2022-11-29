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
	"strconv"

	"github.com/tasselsd/umeq-csi/pkg/state"
)

type Attacher interface {
	Attach(nodeId, volumeId, qcow2Path string) error
	Detach(nodeId, volumeId string) error
	DevPath(volumeId string) (string, error)
	Clean(volumeId string) error
}

type CommonAttacher struct {
	kv state.KvStore
}

func (a *CommonAttacher) nextSeq() string {
	a.kv.Lock("global")
	defer a.kv.Unlock("global")
	r, err := a.kv.Get("/xiaomakai/id")
	if err != nil {
		a.kv.Set("/xiaomakai/id", []byte("1"))
		return "0"
	}
	seq := string(r)
	val, err := strconv.ParseInt(seq, 10, 64)
	if err != nil {
		panic(err)
	}
	a.kv.Set("/xiaomakai/id", []byte(fmt.Sprintf("%d", val+1)))
	return seq
}

func (a *CommonAttacher) getSerialId(volumeId string) (string, error) {
	r, err := a.kv.Get("/xiaomakai/" + volumeId)
	if err != nil {
		id := a.nextSeq()
		err = a.kv.Set("/xiaomakai/"+volumeId, []byte(id))
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return string(r), nil
}

func (a *CommonAttacher) Clean(volumeId string) error {
	log.Println("[info] clean", volumeId, "attach info")
	return a.kv.Del("/xiaomakai/" + volumeId)
}

func (q *CommonAttacher) DevPath(volumeId string) (string, error) {
	r, err := q.kv.Get("/xiaomakai/" + volumeId)
	if err != nil {
		return "", fmt.Errorf("volume %s not found! not attach yet?", volumeId)
	}
	return "/dev/disk/by-id/virtio-" + string(r), nil
}
