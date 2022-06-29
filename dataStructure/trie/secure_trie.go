package trie

import (
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
)

var secureKeyPrefix = []byte("secure-key-")

const secureKeyLength = 11 + 32

type SecureTrie struct {
	trie             Trie
	hashKeyBuf       [secureKeyLength]byte
	secKeyBuf        [200]byte
	secKeyCache      map[string][]byte
	secKeyCacheOwner *SecureTrie
}

func NewSecure(root common.Hash, db kvDatabase.IDriver, cacheLimit uint16) (*SecureTrie, error) {
	if db == nil {
		panic("NewSecure called with nil database")
	}
	trie, err := New(root, db)
	if err != nil {
		return nil, err
	}
	trie.SetCacheLimit(cacheLimit)
	return &SecureTrie{trie: *trie}, nil
}

func (t *SecureTrie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		log.Log.Errorf("Unhandled trie error: %v", err)
	}
	return res
}

func (t *SecureTrie) TryGet(key []byte) ([]byte, error) {
	return t.trie.TryGet(t.hashKey(key))
}

func (t *SecureTrie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		log.Log.Errorf("Unhandled trie error: %v", err)
	}
}

func (t *SecureTrie) TryUpdate(key, value []byte) error {
	hk := t.hashKey(key)
	err := t.trie.TryUpdate(hk, value)
	if err != nil {
		return err
	}
	t.getSecKeyCache()[string(hk)] = common.CopyBytes(key)
	return nil
}

func (t *SecureTrie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil {
		log.Log.Errorf("Unhandled trie error: %v", err)
	}
}

func (t *SecureTrie) TryDelete(key []byte) error {
	hk := t.hashKey(key)
	delete(t.getSecKeyCache(), string(hk))
	return t.trie.TryDelete(hk)
}

func (t *SecureTrie) GetKey(shaKey []byte) []byte {
	if key, ok := t.getSecKeyCache()[string(shaKey)]; ok {
		return key
	}
	key, _ := t.trie.db.Get(t.secKey(shaKey))
	return key.Data
}

func (t *SecureTrie) Commit() (root common.Hash, err error) {
	return t.CommitTo(t.trie.db)
}

func (t *SecureTrie) Hash() common.Hash {
	return t.trie.Hash()
}

func (t *SecureTrie) Root() []byte {
	return t.trie.Root()
}

func (t *SecureTrie) Copy() *SecureTrie {
	cpy := *t
	return &cpy
}

func (t *SecureTrie) NodeIterator(start []byte) NodeIterator {
	return t.trie.NodeIterator(start)
}

func (t *SecureTrie) CommitTo(db kvDatabase.IDriver) (root common.Hash, err error) {
	if len(t.getSecKeyCache()) > 0 {
		for hk, key := range t.secKeyCache {
			if err := db.Put(kvDatabase.KVItem{
				Key:  t.secKey([]byte(hk)),
				Data: key,
			}); err != nil {
				return common.Hash{}, err
			}
		}
		t.secKeyCache = make(map[string][]byte)
	}
	return t.trie.CommitTo(db)
}

func (t *SecureTrie) secKey(key []byte) []byte {
	buf := append(t.secKeyBuf[:0], secureKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

func (t *SecureTrie) hashKey(key []byte) []byte {
	h := newHasher(0, 0)
	h.sha.Reset()
	h.sha.Write(key)
	buf := h.sha.Sum(t.hashKeyBuf[:0])
	returnHasherToPool(h)
	return buf
}

func (t *SecureTrie) getSecKeyCache() map[string][]byte {
	if t != t.secKeyCacheOwner {
		t.secKeyCacheOwner = t
		t.secKeyCache = make(map[string][]byte)
	}
	return t.secKeyCache
}
