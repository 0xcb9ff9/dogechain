package kvstorage

import (
	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/hashicorp/go-hclog"
)

type kvStorageBuilder struct {
	logger        hclog.Logger
	badgerBuilder kvdb.BadgerBuilder
}

func (builder *kvStorageBuilder) Build() (storage.Storage, error) {
	db, err := builder.badgerBuilder.Build()
	if err != nil {
		return nil, err
	}

	return newKeyValueStorage(builder.logger.Named("kvstorage"), db), nil
}

// NewKVStorageBuilder creates the new blockchain storage builder
func NewKVStorageBuilder(logger hclog.Logger, badgerBuilder kvdb.BadgerBuilder) storage.StorageBuilder {
	return &kvStorageBuilder{
		logger:        logger,
		badgerBuilder: badgerBuilder,
	}
}
