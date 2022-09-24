package attach

import (
	"log"

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

func (a *CommonAttacher) Clean(volumeId string) error {
	log.Println("[info] request clean volume")
	return nil
}
