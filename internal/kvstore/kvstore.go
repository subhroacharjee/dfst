package kvstore

type KVStore interface {
	Add(key string, value string)
	Get(key string) (string, bool)
	Delete(key string)
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}
