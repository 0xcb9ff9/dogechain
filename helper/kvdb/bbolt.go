package kvdb

import (
	"encoding/binary"
	"hash/crc32"
	"log"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/atomic"
)

type hashBucketID [2]byte

func (h *hashBucketID) Bytes() []byte {
	return h[:]
}

func newHashBucketID(key []byte) hashBucketID {
	var bucket hashBucketID

	keyCrc := crc32.ChecksumIEEE(key)

	hashKey := make([]byte, 4)
	binary.LittleEndian.PutUint32(hashKey, keyCrc)

	copy(bucket[:], hashKey[:2])

	return bucket
}

// boltKV is the bolt implementation of the kv storage
type boltKV struct {
	db        *bolt.DB
	closeFlag atomic.Bool
}

type boltBatch struct {
	tx        *bolt.Tx
	bucketMap map[hashBucketID]*bolt.Bucket
}

func (b *boltBatch) Set(k, v []byte) error {
	keyBucketID := newHashBucketID(k)

	if b.bucketMap[keyBucketID] == nil {
		var err error
		b.bucketMap[keyBucketID], err = b.tx.CreateBucketIfNotExists(keyBucketID.Bytes())

		if err != nil {
			return err
		}
	}

	bucket := b.bucketMap[keyBucketID]

	if err := bucket.Put(k, v); err != nil {
		return err
	}

	return nil
}

func (b *boltBatch) Get(k []byte) ([]byte, bool, error) {
	keyBucketID := newHashBucketID(k)

	if b.bucketMap[keyBucketID] == nil {
		var err error
		b.bucketMap[keyBucketID], err = b.tx.CreateBucketIfNotExists(keyBucketID.Bytes())

		if err != nil {
			return nil, false, err
		}
	}

	bucket := b.bucketMap[keyBucketID]

	val := bucket.Get(k)
	if val != nil {
		valCopy := make([]byte, len(val))
		copy(valCopy, val)

		return valCopy, true, nil
	}

	return nil, false, nil
}

func (b *boltBatch) Write() error {
	return b.tx.Commit()
}

func (b *boltBatch) Close() error {
	return b.tx.Rollback()
}

func (kv *boltKV) Batch() (KVBatch, error) {
	tx, err := kv.db.Begin(true)
	if err != nil {
		return nil, err
	}

	return &boltBatch{
		tx:        tx,
		bucketMap: make(map[hashBucketID]*bolt.Bucket),
	}, nil
}

// Set sets the key-value pair in bolt storage
func (kv *boltKV) Set(k []byte, v []byte) error {
	err := kv.db.Update(func(tx *bolt.Tx) error {
		bucketID := newHashBucketID(k)
		bucket, err := tx.CreateBucketIfNotExists(bucketID.Bytes())
		if err != nil {
			return err
		}

		return bucket.Put(k, v)
	})

	if err != nil {
		log.Println(err)
	}

	return err
}

// Get retrieves the key-value pair in bolt storage
func (kv *boltKV) Get(k []byte) ([]byte, bool, error) {
	var valCopy []byte = nil

	err := kv.db.View(func(tx *bolt.Tx) error {
		bucketID := newHashBucketID(k)
		bucket := tx.Bucket(bucketID.Bytes())
		if bucket == nil {
			return nil
		}

		val := bucket.Get(k)
		if val != nil {
			valCopy = make([]byte, len(val))
			copy(valCopy, val)
		}

		return nil
	})

	if valCopy == nil {
		return nil, false, nil
	}

	if err != nil {
		log.Println(err)

		return nil, false, err
	}

	return valCopy, true, nil
}

// Close closes the bolt storage instance
func (kv *boltKV) Close() error {
	if kv.closeFlag.Load() {
		return nil
	}

	kv.closeFlag.Store(true)
	err := kv.db.Sync()

	if err != nil {
		log.Println(err)
	}

	return kv.db.Close()
}
