package kvdb

import (
	"errors"
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

// badgerKV is the leveldb implementation of the kv storage
type badgerKV struct {
	db        *badger.DB
	closeFlag bool
	closeCh   chan struct{}
}

type batchKV struct {
	k, v []byte
}

type badgerBatch struct {
	kvPairs []batchKV
	db      *badger.DB
}

func (b *badgerBatch) Set(k, v []byte) {
	b.kvPairs = append(b.kvPairs, batchKV{k: k, v: v})
}

func (b *badgerBatch) Write() error {
	batch := b.db.NewWriteBatch()
	defer batch.Cancel()

	batch.SetMaxPendingTxns(len(b.kvPairs) + 1)

	for _, pair := range b.kvPairs {
		err := batch.SetEntry(badger.NewEntry(pair.k, pair.v))
		if err != nil {
			log.Panicln(err)

			return err
		}
	}

	err := batch.Flush()

	if err != nil {
		log.Println(err)
	}

	if err != nil {
		log.Println(err)
	}

	return err
}

func (kv *badgerKV) Batch() KVBatch {
	return &badgerBatch{db: kv.db, kvPairs: make([]batchKV, 0)}
}

// Set sets the key-value pair in leveldb storage
func (kv *badgerKV) Set(k []byte, v []byte) error {
	err := kv.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})

	if err != nil {
		log.Println(err)
	}

	return err
}

// Get retrieves the key-value pair in leveldb storage
func (kv *badgerKV) Get(k []byte) ([]byte, bool, error) {
	var valCopy []byte

	err := kv.db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err := txn.Get(k)

		if err == nil {
			valCopy, err = item.ValueCopy(nil)
		}

		return err
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, false, nil
	}

	if err != nil {
		log.Println(err)

		return nil, false, err
	}

	return valCopy, true, nil
}

func (kv *badgerKV) Sync() error {
	return kv.db.Sync()
}

// Close closes the leveldb storage instance
func (kv *badgerKV) Close() error {
	if kv.closeFlag {
		return nil
	}

	kv.closeFlag = true

	close(kv.closeCh)

	err := kv.db.Sync()

	if err != nil {
		log.Println(err)
	}

	return kv.db.Close()
}
