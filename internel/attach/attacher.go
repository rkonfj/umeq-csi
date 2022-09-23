package attach

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Attacher interface {
	Attach(nodeId, volumeId, qcow2Path string) error
	Detach(nodeId, volumeId string) error
	DevPath(volumeId string) (string, error)
	Clean(volumeId string) error
}

type CommonAttacher struct {
	etcdctl *clientv3.Client
}

func (a *CommonAttacher) lock(nodeId string) error {
	s, _ := concurrency.NewSession(a.etcdctl)
	defer s.Close()
	l := concurrency.NewMutex(s, nodeId)
	ctx := context.Background()
	if err := l.Lock(ctx); err != nil {
		return err
	}
	return nil
}

func (a *CommonAttacher) unlock(nodeId string) error {
	s, _ := concurrency.NewSession(a.etcdctl)
	defer s.Close()
	l := concurrency.NewMutex(s, nodeId)
	ctx := context.Background()
	return l.Unlock(ctx)
}

func (a *CommonAttacher) Clean(volumeId string) error {
	log.Println("[info] request clean volume")
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := a.etcdctl.Delete(c, "/xiaomakai/"+volumeId)
	if err != nil {
		log.Println("[etcd] delete ERR:", err)
		return err
	}
	log.Printf("[etcd] delete successfully! deleted: %d\n", resp.Deleted)
	return nil
}
