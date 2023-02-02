package itrie

import (
	"bytes"
	"errors"

	"github.com/dogechain-lab/dogechain/crypto"
	"github.com/dogechain-lab/dogechain/state"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/dogechain-lab/fastrlp"
)

type Trie struct {
	stateDB StateDB
	root    Node

	epoch uint32
}

func NewTrie() *Trie {
	return &Trie{}
}

func (t *Trie) Get(k []byte) ([]byte, bool) {
	txn := t.Txn()

	res, err := txn.Lookup(k)
	if err != nil {
		// maby return error? interface need changed, big change
		t.stateDB.Logger().Error("Failed to lookup key", "key", k, "err", err)
	}

	return res, res != nil
}

func hashit(k []byte) []byte {
	return crypto.Keccak256(k)
}

var accountArenaPool fastrlp.ArenaPool

var stateArenaPool fastrlp.ArenaPool // TODO, Remove once we do update in fastrlp

func (t *Trie) Commit(objs []*state.Object) (state.Snapshot, []byte, error) {
	var root []byte = nil

	var nTrie *Trie = nil

	metrics := t.stateDB.GetMetrics()

	// metrics logger
	insertCount := 0
	deleteCount := 0
	newSetCodeCount := 0

	// Create an insertion batch for all the entries
	err := t.stateDB.Transaction(func(st StateDBTransaction) error {
		defer st.Rollback()

		tt := t.Txn()
		tt.reader = st

		arena := accountArenaPool.Get()
		defer accountArenaPool.Put(arena)

		ar1 := stateArenaPool.Get()
		defer stateArenaPool.Put(ar1)

		for _, obj := range objs {
			if obj.Deleted {
				err := tt.Delete(hashit(obj.Address.Bytes()))
				if err != nil {
					return err
				}

				deleteCount++
			} else {
				account := state.Account{
					Balance:  obj.Balance,
					Nonce:    obj.Nonce,
					CodeHash: obj.CodeHash.Bytes(),
					Root:     obj.Root, // old root
				}

				if len(obj.Storage) != 0 {
					localSnapshot, err := t.stateDB.NewSnapshotAt(obj.Root)
					if err != nil {
						return err
					}

					trie, ok := localSnapshot.(*Trie)
					if !ok {
						return errors.New("invalid type assertion")
					}

					localTxn := trie.Txn()

					for _, entry := range obj.Storage {
						k := hashit(entry.Key)
						if entry.Deleted {
							err := localTxn.Delete(k)
							if err != nil {
								return err
							}

							deleteCount++
						} else {
							vv := ar1.NewBytes(bytes.TrimLeft(entry.Val, "\x00"))
							err := localTxn.Insert(k, vv.MarshalTo(nil))
							if err != nil {
								return err
							}

							insertCount++
						}
					}

					// observe account hash time
					observe := metrics.transactionAccountHashSecondsObserve()

					accountStateRoot, _ := localTxn.Hash(st)

					// end observe account hash time
					observe()

					account.Root = types.BytesToHash(accountStateRoot)
				}

				if obj.DirtyCode {
					// write code to memory object, never failed
					// if failed, can't alloc memory, it will panic
					err := st.SetCode(obj.CodeHash, obj.Code)
					if err != nil {
						return err
					}

					newSetCodeCount++
				}

				vv := account.MarshalWith(arena)
				data := vv.MarshalTo(nil)

				tt.Insert(hashit(obj.Address.Bytes()), data)
				insertCount++

				arena.Reset()
			}
		}

		var err error

		// observe root hash time
		observe := metrics.transactionAccountHashSecondsObserve()

		root, err = tt.Hash(st)
		if err != nil {
			return err
		}

		// end observe root hash time
		observe()

		// dont use st here, we need to use the original stateDB
		nTrie = NewTrie()
		nTrie.stateDB = t.stateDB
		nTrie.root = tt.root
		nTrie.epoch = tt.epoch

		// Commit all the entries to db
		return st.Commit()
	})

	if err == nil {
		metrics.transactionInsertObserve(insertCount)
		metrics.transactionDeleteObserve(deleteCount)
		metrics.transactionNewAccountObserve(newSetCodeCount)
	}

	return nTrie, root, err
}

// Hash returns the root hash of the trie. It does not write to the
// database and can be used even if the trie doesn't have one.
func (t *Trie) Hash() types.Hash {
	if t.root == nil {
		return types.EmptyRootHash
	}

	hash, cached, _ := t.hashRoot()
	t.root = cached

	return types.BytesToHash(hash)
}

func (t *Trie) hashRoot() ([]byte, Node, error) {
	hash, _ := t.root.Hash()

	return hash, t.root, nil
}

func (t *Trie) Txn() *Txn {
	return &Txn{reader: t.stateDB, root: t.root, epoch: t.epoch + 1}
}

func prefixLen(k1, k2 []byte) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}

	var i int

	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}

	return i
}
