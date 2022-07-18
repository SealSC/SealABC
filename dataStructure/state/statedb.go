package state

import (
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/dataStructure/trie"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"

	"github.com/ethereum/go-ethereum/rlp"

	"math/big"
	"sync"
)

type StateDB struct {
	db   Database
	trie Trie

	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	dbErr error

	tHash, bHash common.Hash
	txIndex      int

	lock sync.Mutex

	accountTool AccountTool

	cryptoTools crypto.Tools
}

func New(root common.Hash, db Database, cryptoTools crypto.Tools, accountTool AccountTool) (*StateDB, error) {
	tr, err := db.OpenTrie(root, cryptoTools)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		cryptoTools:       cryptoTools,
		accountTool:       accountTool,
	}, nil
}

func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *StateDB) Error() error {
	return s.dbErr
}

func (s *StateDB) Reset(root common.Hash) error {
	tr, err := s.db.OpenTrie(root, s.cryptoTools)
	if err != nil {
		return err
	}
	s.trie = tr
	s.stateObjects = make(map[common.Address]*stateObject)
	s.stateObjectsDirty = make(map[common.Address]struct{})
	s.tHash = common.Hash{}
	s.bHash = common.Hash{}
	s.txIndex = 0
	return nil
}

func (s *StateDB) Exist(addr common.Address) bool {
	return s.getStateObject(addr) != nil
}

func (s *StateDB) Empty(addr common.Address) bool {
	so := s.getStateObject(addr)
	return so == nil || so.empty()
}

func (s *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}

func (s *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

func (s *StateDB) GetCode(addr common.Address) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code(s.db)
	}
	return nil
}

func (s *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	size, err := s.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		s.setError(err)
	}
	return size
}

func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (s *StateDB) GetState(a common.Address, b common.Hash) common.Hash {
	stateObject := s.getStateObject(a)
	if stateObject != nil {
		return stateObject.GetState(s.db, b)
	}
	return common.Hash{}
}

func (s *StateDB) StorageTrie(a common.Address) Trie {
	stateObject := s.getStateObject(a)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(s, nil)
	return cpy.updateTrie(s.db)
}

func (s *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

func (s *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (s *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(common.BytesToHash(s.cryptoTools.HashCalculator.Sum(code)), code)
	}
}

func (s *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(s.db, key, value)
	}
}

func (s *StateDB) Suicide(addr common.Address) bool {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	stateObject.markSuicided()
	stateObject.data.SetBalance(new(big.Int))

	return true
}

func (s *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	s.setError(s.trie.TryUpdate(addr[:], data))
}

func (s *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	s.setError(s.trie.TryDelete(addr[:]))
}

func (s *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	if obj := s.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	enc, err := s.trie.TryGet(addr[:])
	if len(enc) == 0 {
		s.setError(err)
		return nil
	}

	account, err := s.accountTool.DecodeAccount(enc)
	if err != nil {
		log.Log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}

	// Insert into the live set.
	obj := newObject(s, s.cryptoTools, addr, account, s.MarkStateObjectDirty)
	s.setStateObject(obj)
	return obj
}

func (s *StateDB) setStateObject(object *stateObject) {
	s.stateObjects[object.Address()] = object
}

func (s *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := s.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = s.createObject(addr)
	}
	return stateObject
}

func (s *StateDB) MarkStateObjectDirty(addr common.Address) {
	s.stateObjectsDirty[addr] = struct{}{}
}

func (s *StateDB) createObject(addr common.Address) (newObj, prev *stateObject) {
	prev = s.getStateObject(addr)
	newObj = newObject(s, s.cryptoTools, addr, s.accountTool.NewAccount(), s.MarkStateObjectDirty)
	newObj.setNonce(0)
	s.setStateObject(newObj)
	return newObj, prev
}

func (s *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {
	so := s.getStateObject(addr)
	if so == nil {
		return
	}

	for h, value := range so.cachedStorage {
		cb(h, value)
	}

	it := trie.NewIterator(so.getTrie(s.db).NodeIterator(nil))
	for it.Next() {
		key := common.BytesToHash(s.trie.GetKey(it.Key))
		if _, ok := so.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}

func (s *StateDB) Copy() *StateDB {
	s.lock.Lock()
	defer s.lock.Unlock()

	state := &StateDB{
		db:                s.db,
		trie:              s.trie,
		stateObjects:      make(map[common.Address]*stateObject, len(s.stateObjectsDirty)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(s.stateObjectsDirty)),
		cryptoTools:       s.cryptoTools,
		accountTool:       s.accountTool,
	}
	for addr := range s.stateObjectsDirty {
		state.stateObjects[addr] = s.stateObjects[addr].deepCopy(state, state.MarkStateObjectDirty)
		state.stateObjectsDirty[addr] = struct{}{}
	}
	return state
}

func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range s.stateObjectsDirty {
		stateObject := s.stateObjects[addr]
		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
	}
}

func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.Finalise(deleteEmptyObjects)
	return s.trie.Hash()
}

func (s *StateDB) Prepare(tHash, bHash common.Hash, ti int) {
	s.tHash = tHash
	s.bHash = bHash
	s.txIndex = ti
}

func (s *StateDB) DeleteSuicides() {

	for addr := range s.stateObjectsDirty {
		stateObject := s.stateObjects[addr]

		if stateObject.suicided {
			stateObject.deleted = true
		}
		delete(s.stateObjectsDirty, addr)
	}
}

func (s *StateDB) CommitTo(db kvDatabase.IDriver, deleteEmptyObjects bool) (root common.Hash, err error) {
	var kvList []kvDatabase.KVItem
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			s.deleteStateObject(stateObject)
		case isDirty:
			if stateObject.code != nil && stateObject.dirtyCode {
				kvList = append(kvList, kvDatabase.KVItem{
					Key:  stateObject.CodeHash(),
					Data: stateObject.code,
				})
				stateObject.dirtyCode = false
			}
			if err := stateObject.CommitTrie(s.db, db); err != nil {
				return common.Hash{}, err
			}
			s.updateStateObject(stateObject)
		}

		delete(s.stateObjectsDirty, addr)
	}

	err = db.BatchPut(kvList)
	if err != nil {
		return common.Hash{}, err
	}

	root, err = s.trie.CommitTo(db)

	return root, err
}
