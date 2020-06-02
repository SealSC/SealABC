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
	"SealABC/dataStructure/enum"
	"SealEVM/environment"
	"SealEVM/evmInt256"
)

const ContractAddressLen = 24
func ContractAddressPrefix() []byte {
	return []byte("SC:")
}

type prefixEl struct {
	enum.Element
}

func (p prefixEl) BuildKey(baseKey []byte, extra ...[]byte) []byte {
	keyPrefix := []byte(p.String())

	if baseKey == nil {
		return keyPrefix
	}

	result := append(keyPrefix, baseKey...)
	for _, k := range extra {
		result = append(result, k...)
	}

	return result
}

var StoragePrefixes struct {
	Assets       prefixEl
	Transaction  prefixEl
	Balance      prefixEl

	ContractLog       prefixEl
	ContractData      prefixEl
	ContractCode      prefixEl
	ContractHash      prefixEl
	ContractDestructs prefixEl
}

func (l Ledger) getTxFromStorage(hash []byte) (tx *Transaction, exists bool, err error) {
	key := StoragePrefixes.Transaction.BuildKey(hash)

	txJson, err := l.Storage.Get(key)
	if err != nil {
		return
	}

	exists = txJson.Exists
	if exists {
		tx = &Transaction{}
		err = structSerializer.FromMFBytes(txJson.Data, tx)
	}

	return
}

type contractStorage struct {
	basedLedger *Ledger
}

func (c *contractStorage) GetBalance(address *evmInt256.Int) (*evmInt256.Int, error) {
	balance, err := c.basedLedger.balanceOf(address.Bytes(), c.basedLedger.genesisAssets.getHash())

	var ret *evmInt256.Int
	if err == nil {
		ret = evmInt256.FromBigInt(balance)
	}
	return ret, err
}

func (c *contractStorage) CanTransfer(from, to, val *evmInt256.Int) bool {
	balance, err := c.basedLedger.balanceOf(from.Bytes(), c.basedLedger.genesisAssets.getHash())
	if err != nil {
		return false
	}

	return balance.Cmp(val.Int) >= 0
}

func (c *contractStorage) GetCode(address *evmInt256.Int) ([]byte, error) {
	key := StoragePrefixes.ContractCode.BuildKey(address.Bytes())

	codeKV, err := c.basedLedger.Storage.Get(key)
	if err != nil {
		return nil, err
	}

	return codeKV.Data, nil
}

func (c *contractStorage) GetCodeSize(address *evmInt256.Int) (*evmInt256.Int, error) {
	key := StoragePrefixes.ContractCode.BuildKey(address.Bytes())

	codeKV, err := c.basedLedger.Storage.Get(key)
	if err != nil {
		return nil, err
	}

	if codeKV.Exists {
		return evmInt256.New(int64(len(codeKV.Data))), nil
	}
	return evmInt256.New(0), nil
}

func (c *contractStorage) GetCodeHash(address *evmInt256.Int) (*evmInt256.Int, error) {
	key := StoragePrefixes.ContractHash.BuildKey(address.Bytes())
	hashKV, err := c.basedLedger.Storage.Get(key)
	if err != nil {
		return nil, err
	}

	ret := evmInt256.New(0)
	//todo: if not exists, in ethereum protocol there's several return situations need to be done in future
	if hashKV.Exists {
		ret.SetBytes(hashKV.Data)
	}
	return ret, nil
}

func (c *contractStorage) GetBlockHash(block *evmInt256.Int) (*evmInt256.Int, error) {
	blk, err := c.basedLedger.chain.GetBlockByHeight(block.Uint64())
	if err != nil {
		return nil, err
	}

	hash := evmInt256.New(0)
	hash.SetBytes(blk.Seal.Hash)
	return hash, nil
}

func (c *contractStorage) CreateAddress(caller *evmInt256.Int, tx environment.Transaction) *evmInt256.Int {
	//in seal abc smart assets application, we always create fixed contract address.
	return c.CreateFixedAddress(caller, nil, tx)
}

func (c *contractStorage) CreateFixedAddress(caller *evmInt256.Int, salt *evmInt256.Int, tx environment.Transaction) *evmInt256.Int {
	addrPrefix := ContractAddressPrefix()

	hashCalc := c.basedLedger.CryptoTools.HashCalculator

	baseBytes := caller.Bytes()
	baseBytes = append(baseBytes, tx.TxHash...)

	if salt != nil {
		baseBytes = append(baseBytes, salt.Bytes()...)
	}

	addrHashBytes := hashCalc.Sum(baseBytes)
	orgHashLen := len(addrHashBytes)
	if orgHashLen > ContractAddressLen {
		addrHashBytes = addrHashBytes[orgHashLen-ContractAddressLen:]
	} else if orgHashLen < ContractAddressLen {
		paddingHashBytes := make([]byte, ContractAddressLen, ContractAddressLen)
		copy(paddingHashBytes, addrHashBytes)
		addrHashBytes = paddingHashBytes
	}

	addrBytes := append(addrPrefix, addrHashBytes...)

	ret := evmInt256.New(0)
	ret.SetBytes(addrBytes)
	return ret
}

func (c *contractStorage) Load(n string, k string) (*evmInt256.Int, error) {
	key := StoragePrefixes.ContractData.BuildKey([]byte(n), []byte(k))
	data, err := c.basedLedger.Storage.Get(key)

	if err != nil {
		return nil, err
	}

	ret := evmInt256.New(0)
	if data.Exists {
		ret.SetBytes(data.Data)
	}

	return ret, nil
}
