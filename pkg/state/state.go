package state

type KV struct {
	CodedKey string
	Key      string
	Value    []byte
}

type KvStore interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error
	Lock(key string) error
	Unlock(key string) error
	List() ([]*KV, error)
}
