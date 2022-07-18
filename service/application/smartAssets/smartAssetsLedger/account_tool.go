package smartAssetsLedger

import (
	"github.com/SealSC/SealABC/dataStructure/state"
	"github.com/SealSC/SealABC/log"
	"github.com/ethereum/go-ethereum/rlp"
)

type AccountTool struct {
}

func (a AccountTool) NewAccount() state.Account {
	return &Account{}
}

func (a AccountTool) DecodeAccount(enc []byte) (state.Account, error) {
	var account Account
	if err := rlp.DecodeBytes(enc, &account); err != nil {
		log.Log.Error("Failed to decode state object", "err", err)
		return nil, err
	}
	return &account, nil
}
