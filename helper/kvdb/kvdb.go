package kvdb

type KVBatch interface {
	Set(k, v []byte) error
	Get(k []byte) ([]byte, bool, error)

	Write() error
	Close() error
}

// KVStorage is a k/v storage on memory or bolt
type KVStorage interface {
	Set(k, v []byte) error
	Get(k []byte) ([]byte, bool, error)

	Close() error
}

// KVBatchStorage is a batch write for bolt
type KVBatchStorage interface {
	KVStorage
	Batch() (KVBatch, error)
}
