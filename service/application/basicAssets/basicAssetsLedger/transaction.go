/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package basicAssetsLedger

import (
	"encoding/hex"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/seal"
)

var TransactionTypes struct {
	IssueAssets    enum.Element
	Transfer       enum.Element
	IncreaseSupply enum.Element

	StartSelling enum.Element
	StopSelling  enum.Element
	BuyAssets    enum.Element
}

type SellingData struct {
	Price         uint64 `json:",string"`
	Amount        uint64 `json:",string"`
	Seller        []byte
	SellingAssets []byte
	PaymentAssets []byte
	Transaction   []byte
}

type TransactionData struct {
	TxType string
	Assets Assets
	Memo   string

	Input  []UTXOInput
	Output []UTXOOutput

	ExtraData []byte
}

type Transaction struct {
	TransactionData

	CreateTime int64
	Seal       seal.Entity
}

type TransactionWithBlockInfo struct {
	Transaction
	BlockInfo struct {
		RequestHash []byte
		BlockHeight uint64
		ActionIndex uint32
	}
}

func (t *Transaction) HashString() string {
	return hex.EncodeToString(t.Seal.Hash)
}

func (t *Transaction) Verify(tools crypto.Tools) (err error) {
	txBytes, err := structSerializer.ToMFBytes(t.TransactionData)
	if err != nil {
		return
	}

	_, err = t.Seal.Verify(txBytes, tools.HashCalculator)
	return
}
