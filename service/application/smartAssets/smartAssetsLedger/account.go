package smartAssetsLedger

import (
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"math/big"
)

type Account struct {
	AccountNonce    uint64
	AccountBalance  string
	AccountRoot     []byte
	AccountCodeHash []byte
}

func (a *Account) Balance() *big.Int {
	b := new(big.Int)
	b.SetString(a.AccountBalance, 10)
	return b
}

func (a *Account) SetBalance(b *big.Int) {
	a.AccountBalance = b.String()
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
	return common.BytesToHash(a.AccountRoot)
}

func (a *Account) SetRoot(hash common.Hash) {
	a.AccountRoot = hash.Bytes()
}

func (a *Account) Encode() (data []byte, err error) {
	data, err = structSerializer.ToMFBytes(*a)
	return
}
