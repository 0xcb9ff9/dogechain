package kvstorage

import (
	"os"
	"path"
	"testing"

	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/hashicorp/go-hclog"
)

func newBoltStorage(t *testing.T) (storage.Storage, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("/tmp", "minimal_storage")
	if err != nil {
		t.Fatal(err)
	}

	logger := hclog.NewNullLogger()

	s, err := NewKVStorageBuilder(
		logger, kvdb.NewBoltBuilder(logger, path.Join(tmpDir, "database.db"))).Build()
	if err != nil {
		t.Fatal(err)
	}

	closeFn := func() {
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}

	return s, closeFn
}

func TestBoltStorage(t *testing.T) {
	storage.TestStorage(t, newBoltStorage)
}
