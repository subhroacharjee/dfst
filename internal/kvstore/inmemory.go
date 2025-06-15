package kvstore

import (
	"encoding/json"
	"sync"
)

// Thread safe in memory key value store
type InMemoryStore struct {
	kv map[string]string

	mu sync.RWMutex
}

func NewInMemoryStore() KVStore {
	return &InMemoryStore{
		kv: make(map[string]string),
	}
}

func (k *InMemoryStore) Add(key string, value string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.kv[key] = value
}

func (k *InMemoryStore) Get(key string) (string, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	val, ok := k.kv[key]
	return val, ok
}

func (k *InMemoryStore) Delete(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	delete(k.kv, key)
}

func (k *InMemoryStore) Marshal() ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return json.Marshal(k.kv)
}

func (k *InMemoryStore) Unmarshal(data []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	return json.Unmarshal(data, &k.kv)
}
