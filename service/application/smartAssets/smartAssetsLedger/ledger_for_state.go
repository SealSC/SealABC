package smartAssetsLedger

import (
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/state"
	"math/big"
)

func (l *Ledger) NewState(root []byte) (stateDB *state.StateDB, err error) {
	stateRoot := common.Hash{}
	if len(root) != 0 {
		stateRoot = common.BytesToHash(root)
	}

	stateDB, err = state.New(stateRoot, state.NewDatabase(l.Storage), l.CryptoTools, &AccountTool{})
	if err != nil {
		err = fmt.Errorf("failed to create state trie at %x: %v", root, err)
	}
	return
}

func (l *Ledger) BalanceOf(address []byte) (balance *big.Int, err error) {
	balance = l.stateDB.GetBalance(common.BytesToAddress(address))
	return
}

func (l *Ledger) NonceOf(address []byte) (nonce uint64) {
	nonce = l.stateDB.GetNonce(common.BytesToAddress(address))
	return
}

func (l *Ledger) CodeHashOf(address []byte) (hash []byte) {
	hash = l.stateDB.GetCodeHash(common.BytesToAddress(address)).Bytes()
	return
}

func (l *Ledger) CodeSizeOf(address []byte) (size int) {
	size = l.stateDB.GetCodeSize(common.BytesToAddress(address))
	return
}

func (l *Ledger) CodeOf(address []byte) []byte {
	return l.stateDB.GetCode(common.BytesToAddress(address))
}

func (l *Ledger) StateOf(address []byte, key []byte) []byte {
	return l.stateDB.GetState(common.BytesToAddress(address), common.BytesToHash(key)).Bytes()
}
