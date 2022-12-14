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

package state

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FsKvStore struct {
	root string
	l    sync.Locker
	lMap map[string]sync.Locker
}

func NewFsKvStore(root string) KvStore {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(root, 0600)
		if err != nil {
			panic(err)
		}
	}

	lMap := make(map[string]sync.Locker)
	kv := &FsKvStore{
		root: root,
		l:    &sync.Mutex{},
		lMap: lMap,
	}

	return kv
}

func (kv *FsKvStore) encode(key string) string {
	dst := make([]byte, hex.EncodedLen(len(key)))
	hex.Encode(dst, []byte(key))
	return string(dst)
}

func (kv *FsKvStore) decode(key string) string {
	dst := make([]byte, hex.DecodedLen(len(key)))
	hex.Decode(dst, []byte(key))
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

func (kv *FsKvStore) List() ([]*KV, error) {
	files, err := os.ReadDir(kv.root)
	if err != nil {
		return nil, err
	}
	var kvs []*KV
	for _, f := range files {
		b, _ := os.ReadFile(filepath.Join(kv.root, f.Name()))
		kvs = append(kvs, &KV{
			CodedKey: f.Name(),
			Key:      kv.decode(f.Name()),
			Value:    b,
		})
	}
	return kvs, nil
}

func (kv *FsKvStore) Lock(key string) error {
	kv.l.Lock()
	if l, ok := kv.lMap[key]; ok {
		kv.l.Unlock()
		l.Lock()
		return nil
	}
	l := &sync.Mutex{}
	kv.lMap[key] = l
	kv.l.Unlock()
	l.Lock()
	return nil
}

func (kv *FsKvStore) Unlock(key string) error {
	kv.l.Lock()
	defer kv.l.Unlock()
	if l, ok := kv.lMap[key]; ok {
		l.Unlock()
		delete(kv.lMap, key)
		return nil
	}
	return nil
}
