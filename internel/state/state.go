package state

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type KvStore interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error
	Lock(key string) error
	Unlock(key string) error
}

type FsKvStore struct {
	root string
	l    sync.Locker
}

func NewFsKvStore(root string) KvStore {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(root, 0600)
		if err != nil {
			panic(err)
		}
	}

	kv := &FsKvStore{
		root: root,
		l:    &sync.Mutex{},
	}

	return kv
}

func (kv *FsKvStore) encode(key string) string {
	dst := make([]byte, hex.EncodedLen(len(key)))
	hex.Encode(dst, []byte(key))
	return string(dst)
}

func (kv *FsKvStore) Set(key string, value []byte) error {
	return os.WriteFile(filepath.Join(kv.root, kv.encode(key)), value, 0600)
}

func (kv *FsKvStore) Get(key string) ([]byte, error) {
	if _, err := os.Stat(filepath.Join(kv.root, kv.encode(key))); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("key %s not found", kv.encode(key))
	}
	return os.ReadFile(filepath.Join(kv.root, kv.encode(key)))
}

func (kv *FsKvStore) Del(key string) error {
	return os.Remove(filepath.Join(kv.root, kv.encode(key)))
}

func (kv *FsKvStore) Lock(key string) error {
	kv.l.Lock()
	return nil
}

func (kv *FsKvStore) Unlock(key string) error {
	kv.l.Unlock()
	return nil
}
