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

package smartAssetsLedger

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/crypto/hashes"
	"SealABC/dataStructure/enum"
	"SealABC/metadata/block"
	"SealABC/metadata/seal"
	"math/big"
)

var TxType struct {
	Transfer       enum.Element
	CreateContract enum.Element
	ContractCall   enum.Element
}

type TransactionData struct {
	Type  string
	From  []byte
	To    []byte
	Value string
	Data  []byte
	SN    []byte
}

type StateData struct {
	Key []byte
	Val []byte
}

type TransactionResult struct {
	Success   bool
	ErrorCode int64
	NewState  []StateData
}

type Transaction struct {
	TransactionData
	TransactionResult

	DataSeal   seal.Entity
}

type TransactionList struct {
	Transactions []Transaction
}

func (t Transaction) getData() []byte {
	data, _ := structSerializer.ToMFBytes(t.Data)
	return data
}

func (t Transaction) toMFBytes() []byte {
	data, _ := structSerializer.ToMFBytes(t)
	return data
}

func (t Transaction) getHash() []byte {
	return t.DataSeal.Hash
}

func (t Transaction) verify(hashCalc hashes.IHashCalculator) (passed bool, err error) {
	passed, err = t.DataSeal.Verify(t.getData(), hashCalc)
	if !passed {
		return
	}

	return
}

func (t Transaction) Execute() (result []byte) {
	switch t.Type {
	case TxType.Transfer.String():

	}

	return
}

type txResultCacheData struct {
	val     *big.Int
	gasLeft uint64
	data    []byte
}

const (
	cachedBlockGasKey = "block gas"
	cachedContractReturnData = "contract return data"
)

type txResultCache map[string] *txResultCacheData
type txPreActuator func(tx Transaction, cache txResultCache, blk block.Entity) (ret []StateData, resultCache txResultCache, err error)
type queryActuator func(req QueryRequest) (ret interface{}, err error)

