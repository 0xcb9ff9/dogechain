package kvdb

import (
	"os"
	"path"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func newTestKVStorage(t *testing.T) KVBatchStorage {
	t.Helper()

	tmpDir, err := os.MkdirTemp("/tmp", "minimal_storage")
	if err != nil {
		t.Fatal(err)
	}

	logger := hclog.NewNullLogger()

	s, err := NewBoltBuilder(logger, path.Join(tmpDir, "database.db")).Build()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	})

	return s
}

func TestKVStorageBatchWrite(t *testing.T) {
	t.Helper()

	storage := newTestKVStorage(t)

	batch, err := storage.Batch()
	if err != nil {
		t.Fatal(err)
	}
	defer batch.Close()

	batch.Set([]byte("key1"), []byte("value1"))
	batch.Set([]byte("key1"), []byte("value2"))

	{
		value, isFind, err := batch.Get([]byte("key1"))
		if err != nil {
			t.Fatal(err)
		}

		if !isFind {
			t.Fatal("key1 not found")
		}

		if string(value) != "value2" {
			t.Fatal("batch value not match")
		}
	}

	batch.Set([]byte("key1"), []byte("value3"))

	err = batch.Write()
	if err != nil {
		t.Fatal(err)
	}

	value, isFind, err := storage.Get([]byte("key1"))
	if err != nil {
		t.Fatal(err)
	}

	if !isFind {
		t.Fatal("key1 not found")
	}

	if string(value) != "value3" {
		t.Fatal("value not equal")
	}
}
