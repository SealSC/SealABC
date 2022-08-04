package state

import (
	"bytes"
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/dataStructure/trie"
	//"github.com/ethereum/go-ethereum/crypto"

	"math/big"
)

var emptyCodeHash []byte

func (c Code) String() string {
	return string(c)
}

type Storage map[common.Hash]common.Hash

func (s Storage) String() (str string) {
	for key, value := range s {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (s Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range s {
		cpy[key] = value
	}

	return cpy
}

type stateObject struct {
	address  common.Address
	addrHash common.Hash
	data     Account
	db       *StateDB

	dbErr error

	trie Trie
	code Code

	cachedStorage Storage // Storage entry cache to avoid duplicate reads
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	dirtyCode bool // true if the code was updated
	suicided  bool
	touched   bool
	deleted   bool
	onDirty   func(addr common.Address) // Callback method to mark a state object newly dirty

	cryptoTools crypto.Tools
}

func newObject(db *StateDB, cryptoTools crypto.Tools, address common.Address, data Account, onDirty func(addr common.Address)) *stateObject {
	emptyCodeHash = cryptoTools.HashCalculator.Sum(nil)

	if data.Balance() == nil {
		data.SetBalance(new(big.Int))
	}
	if data.CodeHash == nil {
		data.SetCodeHash(emptyCodeHash)
	}

	return &stateObject{
		db:            db,
		address:       address,
		addrHash:      common.BytesToHash(cryptoTools.HashCalculator.Sum(address[:])),
		data:          data,
		cachedStorage: make(Storage),
		dirtyStorage:  make(Storage),
		onDirty:       onDirty,
	}
}

func (s *stateObject) EncodeData() ([]byte, error) {
	return s.data.Encode()
}

func (s *stateObject) empty() bool {
	return s.data.Nonce() == 0 && s.data.Balance().Sign() == 0 && bytes.Equal(s.data.CodeHash(), emptyCodeHash)
}

func (s *stateObject) Address() common.Address {
	return s.address
}

func (s *stateObject) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *stateObject) Code(db Database) []byte {
	if s.code != nil {
		return s.code
	}
	if bytes.Equal(s.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.ContractCode(s.addrHash, common.BytesToHash(s.CodeHash()))
	if err != nil {
		s.setError(fmt.Errorf("can't load code hash %x: %v", s.CodeHash(), err))
	}
	s.code = code
	return code
}

func (s *stateObject) SetCode(codeHash common.Hash, code []byte) {
	s.setCode(codeHash, code)
}

func (s *stateObject) setCode(codeHash common.Hash, code []byte) {
	s.code = code
	s.data.SetCodeHash(codeHash[:])
	s.dirtyCode = true
	if s.onDirty != nil {
		s.onDirty(s.Address())
	}
}

func (s *stateObject) SetNonce(nonce uint64) {
	s.setNonce(nonce)
}

func (s *stateObject) setNonce(nonce uint64) {
	s.data.SetNonce(nonce)
	if s.onDirty != nil {
		s.onDirty(s.Address())
	}
}

func (s *stateObject) CodeHash() []byte {
	return s.data.CodeHash()
}

func (s *stateObject) Balance() *big.Int {
	return s.data.Balance()
}

func (s *stateObject) Nonce() uint64 {
	return s.data.Nonce()
}

func (s *stateObject) getTrie(db Database) Trie {
	if s.trie == nil {
		var err error
		s.trie, err = db.OpenStorageTrie(s.addrHash, s.data.Root())
		if err != nil {
			s.trie, _ = db.OpenStorageTrie(s.addrHash, common.Hash{})
			s.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return s.trie
}

func (s *stateObject) GetState(db Database, key common.Hash) common.Hash {
	value, exists := s.cachedStorage[key]
	if exists {
		return value
	}
	enc, err := s.getTrie(db).TryGet(key[:])
	if err != nil {
		s.setError(err)
		return common.Hash{}
	}

	if len(enc) > 0 {
		value.SetBytes(enc)
	}
	if (value != common.Hash{}) {
		s.cachedStorage[key] = value
	}
	return value
}

func (s *stateObject) SetState(db Database, key, value common.Hash) {
	s.setState(key, value)
}

func (s *stateObject) setState(key, value common.Hash) {
	s.cachedStorage[key] = value
	s.dirtyStorage[key] = value

	if s.onDirty != nil {
		s.onDirty(s.Address())
	}
}

func (s *stateObject) updateTrie(db Database) Trie {
	tr := s.getTrie(db)
	for key, value := range s.dirtyStorage {
		delete(s.dirtyStorage, key)
		if (value == common.Hash{}) {
			s.setError(tr.TryDelete(key[:]))
			continue
		}

		s.setError(tr.TryUpdate(key[:], value.Bytes()))
	}
	return tr
}

func (s *stateObject) updateRoot(db Database) {
	s.updateTrie(db)
	s.data.SetRoot(s.trie.Hash())
}

func (s *stateObject) CommitTrie(db Database, bw trie.BatchWriter) error {
	s.updateTrie(db)
	if s.dbErr != nil {
		return s.dbErr
	}
	root, err := s.trie.CommitTo(bw)
	if err == nil {
		s.data.SetRoot(root)
	}
	return err
}

func (s *stateObject) markSuicided() {
	s.suicided = true
	if s.onDirty != nil {
		s.onDirty(s.Address())

	}
}

func (s *stateObject) touch() {
	if s.onDirty != nil {
		s.onDirty(s.Address())
	}
	s.touched = true
}

func (s *stateObject) AddBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		if s.empty() {
			s.touch()
		}

		return
	}
	s.SetBalance(new(big.Int).Add(s.Balance(), amount))
}

func (s *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Sub(s.Balance(), amount))
}

func (s *stateObject) SetBalance(amount *big.Int) {
	s.setBalance(amount)
}

func (s *stateObject) setBalance(amount *big.Int) {
	s.data.SetBalance(amount)
	if s.onDirty != nil {
		s.onDirty(s.Address())
	}
}

func (s *stateObject) ReturnGas(gas *big.Int) {}

func (s *stateObject) deepCopy(db *StateDB, onDirty func(addr common.Address)) *stateObject {
	account := s.deepCopyAccount(db, s.data)
	stateObject := newObject(db, db.cryptoTools, s.address, account, onDirty)
	if s.trie != nil {
		stateObject.trie = db.db.CopyTrie(s.trie)
	}
	stateObject.code = s.code
	stateObject.dirtyStorage = s.dirtyStorage.Copy()
	stateObject.cachedStorage = s.dirtyStorage.Copy()
	stateObject.suicided = s.suicided
	stateObject.dirtyCode = s.dirtyCode
	stateObject.deleted = s.deleted

	return stateObject
}

func (s *stateObject) deepCopyAccount(db *StateDB, data Account) Account {
	account := db.accountTool.NewAccount()
	account.SetRoot(data.Root())
	account.SetBalance(data.Balance())
	account.SetNonce(data.Nonce())
	account.SetCodeHash(data.CodeHash())
	return account
}
