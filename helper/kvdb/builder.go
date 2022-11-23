package kvdb

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
)

const (
	minBadgerIndexCache = 16 // 16 MiB

	DefaultBadgerIndexCache   = 8    // 8 MiB
	DefaultBloomFalsePositive = 0.01 // bloom filter false positive (0.01 = 1%)
	DefaultBaseTablesSize     = 4    // 4 MiB
	DefaultBadgerSyncWrites   = false

	gcTicker = 5 * time.Minute
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func clamp(n, min, max float64) float64 {
	if n < min {
		return min
	}

	if n > max {
		return max
	}

	return n
}

type BadgerBuilder interface {
	// set cache size
	SetCacheSize(int) BadgerBuilder

	// set bloom key bits
	SetBloomFalsePositive(float64) BadgerBuilder

	// set compaction table size
	SetBaseTableSize(int) BadgerBuilder

	// set sync write
	SetSyncWrites(bool) BadgerBuilder

	// build the storage
	Build() (KVBatchStorage, error)
}

type badgerBuilder struct {
	logger  hclog.Logger
	options badger.Options
}

func (builder *badgerBuilder) SetCacheSize(cacheSize int) BadgerBuilder {
	cacheSize = max(cacheSize, minBadgerIndexCache)

	builder.options = builder.options.WithIndexCacheSize(int64(cacheSize) << 20)

	builder.logger.Info("badger",
		"IndexCacheSize", fmt.Sprintf("%d Mib", cacheSize),
	)

	return builder
}

func (builder *badgerBuilder) SetBloomFalsePositive(bloomFalsePositive float64) BadgerBuilder {
	builder.options = builder.options.WithBloomFalsePositive(
		clamp(
			bloomFalsePositive,
			0.0001,
			0.9999,
		))

	builder.logger.Info("badger",
		"Bloom filter False Positive", bloomFalsePositive,
	)

	return builder
}

func (builder *badgerBuilder) SetBaseTableSize(baseTableSize int) BadgerBuilder {
	builder.options = builder.options.WithBaseTableSize(int64(baseTableSize) << 20)

	builder.logger.Info("badger",
		"BaseTableSize", fmt.Sprintf("%d Mib", baseTableSize))

	return builder
}

func (builder *badgerBuilder) SetSyncWrites(sync bool) BadgerBuilder {
	builder.options = builder.options.WithSyncWrites(sync)

	builder.logger.Info("badger",
		"sync", sync,
	)

	return builder
}

func (builder *badgerBuilder) Build() (KVBatchStorage, error) {
	db, err := badger.Open(builder.options)
	if err != nil {
		builder.logger.Error("badger open database", "error", err)

		return nil, err
	}

	closeCh := make(chan struct{})
	ticker := time.NewTicker(gcTicker)

	go func() {
		select {
		case <-ticker.C:
			ticker.Reset(gcTicker)

			err := db.RunValueLogGC(0.7)
			if err != nil {
				builder.logger.Error("badger GC failed", "error", err)
			}
		case <-closeCh:
			ticker.Stop()

			return
		}
	}()

	return &badgerKV{
		db:        db,
		closeFlag: false,
		closeCh:   closeCh,
	}, nil
}

// NewBuilder creates the new leveldb storage builder
func NewBadgerBuilder(logger hclog.Logger, path string) BadgerBuilder {
	return &badgerBuilder{
		logger: logger,
		options: badger.
			DefaultOptions(path).
			WithCompression(options.ZSTD).
			WithIndexCacheSize(DefaultBadgerIndexCache << 20).
			WithBaseTableSize(DefaultBaseTablesSize << 20).
			WithSyncWrites(DefaultBadgerSyncWrites).
			WithMaxLevels(9),
	}
}
