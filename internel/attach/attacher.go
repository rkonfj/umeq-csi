package attach

import (
	"fmt"
	"log"
	"strconv"

	"github.com/tasselsd/umeq-csi/internel/state"
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
