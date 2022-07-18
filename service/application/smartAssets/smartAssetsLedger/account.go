package smartAssetsLedger

import (
	"github.com/SealSC/SealABC/common"
	"math/big"
)

type Account struct {
	AccountNonce    uint64
	AccountBalance  *big.Int
	AccountRoot     common.Hash
	AccountCodeHash []byte
}

func (a *Account) Balance() *big.Int {
	return a.AccountBalance
}

func (a *Account) SetBalance(b *big.Int) {
	a.AccountBalance = b
}

func (a *Account) CodeHash() []byte {
	return a.AccountCodeHash
}

func (a *Account) SetCodeHash(hash []byte) {
	a.AccountCodeHash = hash
}

func (a *Account) Nonce() uint64 {
	return a.AccountNonce
}

func (a *Account) SetNonce(nonce uint64) {
	a.AccountNonce = nonce
}

func (a *Account) Root() common.Hash {
	return a.AccountRoot
}

func (a *Account) SetRoot(hash common.Hash) {
	a.AccountRoot = hash
}
