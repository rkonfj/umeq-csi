package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

type KvStore interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error
	Lock() error
	Unlock() error
}

type FsKvStore struct {
	root string
	l    *flock.Flock
}

func NewFsKvStore(root string) KvStore {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(root, 0600)
		if err != nil {
			panic(err)
		}
	}
	fileLock := flock.New(filepath.Join(root, "x.lock"))
	kv := &FsKvStore{
		root: root,
		l:    fileLock,
	}

	return kv
}

func (kv *FsKvStore) Set(key string, value []byte) error {
	return os.WriteFile(filepath.Join(kv.root, key), value, 0600)
}

func (kv *FsKvStore) Get(key string) ([]byte, error) {
	if _, err := os.Stat(filepath.Join(kv.root, key)); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("key %s not found", key)
	}
	return os.ReadFile(filepath.Join(kv.root, key))
}

func (kv *FsKvStore) Del(key string) error {
	return os.Remove(filepath.Join(kv.root, key))
}

func (kv *FsKvStore) Lock() error {
	return kv.l.Lock()
}

func (kv *FsKvStore) Unlock() error {
	return kv.l.Unlock()
}
