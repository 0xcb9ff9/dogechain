package kvstorage

import (
	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/hashicorp/go-hclog"
)

type boltStorageBuilder struct {
	logger      hclog.Logger
	boltBuilder kvdb.BoltBuilder
}

func (builder *boltStorageBuilder) Build() (storage.Storage, error) {
	db, err := builder.boltBuilder.Build()
	if err != nil {
		return nil, err
	}

	return newKeyValueStorage(builder.logger.Named("bolt"), db), nil
}

// NewboltStorageBuilder creates the new blockchain storage builder
func NewKVStorageBuilder(logger hclog.Logger, boltBuilder kvdb.BoltBuilder) storage.StorageBuilder {
	return &boltStorageBuilder{
		logger:      logger,
		boltBuilder: boltBuilder,
	}
}
