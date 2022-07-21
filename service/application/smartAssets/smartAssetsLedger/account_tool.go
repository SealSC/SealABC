package smartAssetsLedger

import (
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/dataStructure/state"
	"github.com/SealSC/SealABC/log"
)

type AccountTool struct {
}

func (a AccountTool) NewAccount() state.Account {
	return &Account{}
}

func (a AccountTool) DecodeAccount(enc []byte) (state.Account, error) {
	var account Account
	if err := structSerializer.FromMFBytes(enc, &account); err != nil {
		log.Log.Error("Failed to decode state object", "err", err)
		return nil, err
	}
	return &account, nil
}
