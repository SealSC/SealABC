package state

import (
	"github.com/SealSC/SealABC/common"
	"math/big"
)

type Code []byte

type Account interface {
	Balance() *big.Int
	SetBalance(*big.Int)

	CodeHash() []byte
	SetCodeHash([]byte)

	Nonce() uint64
	SetNonce(uint64)

	Root() common.Hash
	SetRoot(common.Hash)

	Encode() ([]byte, error)
}

type AccountTool interface {
	NewAccount() Account
	DecodeAccount([]byte) (Account, error)
}
