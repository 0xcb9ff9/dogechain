package kvdb

import (
	"github.com/hashicorp/go-hclog"
	"go.uber.org/atomic"

	bolt "go.etcd.io/bbolt"
)

const (
	DefaultBoltSyncWrites = true
)

type BoltBuilder interface {
	// set sync write
	SetSyncWrites(bool) BoltBuilder

	// build the storage
	Build() (KVBatchStorage, error)
}

type boltBuilder struct {
	logger hclog.Logger
	path   string
	sync   bool
}

func (builder *boltBuilder) SetSyncWrites(sync bool) BoltBuilder {
	builder.sync = sync

	builder.logger.Info("bolt",
		"sync", sync,
	)

	return builder
}

func (builder *boltBuilder) Build() (KVBatchStorage, error) {
	db, err := bolt.Open(builder.path, 0660, &bolt.Options{
		NoSync: !builder.sync,
	})

	if err != nil {
		return nil, err
	}

	return &boltKV{
		db:        db,
		closeFlag: *atomic.NewBool(false),
	}, nil
}

// NewBuilder creates the new bolt storage builder
func NewBoltBuilder(logger hclog.Logger, path string) BoltBuilder {
	return &boltBuilder{
		logger: logger,
		path:   path,
		sync:   DefaultBoltSyncWrites,
	}
}
