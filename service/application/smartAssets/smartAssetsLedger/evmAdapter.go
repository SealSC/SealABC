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
	"bytes"
	"fmt"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealEVM"
	"github.com/SealSC/SealEVM/common"
	"github.com/SealSC/SealEVM/environment"
	"github.com/SealSC/SealEVM/evmInt256"
	"github.com/SealSC/SealEVM/storage"
	"math/big"
	"sort"
)

const defaultStackDepth = 1000

func constTransactionGasPrice() *evmInt256.Int {
	return evmInt256.New(1)
}
func constTransactionGasLimit() *evmInt256.Int {
	return evmInt256.New(100000000)
}

func (l Ledger) newEVM(tx Transaction, callback SealEVM.EVMResultCallback,
	blk block.Entity, blockGasLimit *evmInt256.Int) (*SealEVM.EVM, *environment.Contract, error) {

	evmTransaction := environment.Transaction{
		TxHash:   tx.DataSeal.Hash,
		Origin:   common.BytesDataToEVMIntHash(tx.DataSeal.Hash),
		GasPrice: constTransactionGasPrice(),
		GasLimit: constTransactionGasLimit(),
	}

	hashByte := l.CryptoTools.HashCalculator.Sum(tx.Data)
	contractHash := common.BytesDataToEVMIntHash(hashByte)
	caller := common.BytesDataToEVMIntHash(tx.DataSeal.SignerPublicKey)
	var contractAddress *evmInt256.Int
	var contractCode []byte
	if len(tx.To) == 0 {
		contractAddress = l.storageForEVM.CreateAddress(caller, evmTransaction)
		contractCode = tx.Data
	} else {
		contractAddress = common.BytesDataToEVMIntHash(tx.To)
		codeData, err := l.storageForEVM.GetCode(contractAddress)
		if err != nil {
			return nil, nil, Errors.ContractNotFound.NewErrorWithNewMessage(err.Error())
		}

		contractCode = codeData
		contractHash, _ = l.storageForEVM.GetCodeHash(contractAddress)
	}

	contract := environment.Contract{
		Namespace: contractAddress,
		Code:      contractCode,
		Hash:      contractHash,
	}

	return SealEVM.New(SealEVM.EVMParam{
		MaxStackDepth:  defaultStackDepth,
		ExternalStore:  &l.storageForEVM,
		ResultCallback: callback,
		Context: &environment.Context{
			Block: environment.Block{
				Coinbase:   common.BytesDataToEVMIntHash(blk.BlankSeal.SignerPublicKey),
				Timestamp:  evmInt256.New(int64(blk.Header.Timestamp)),
				Number:     evmInt256.New(int64(blk.Header.Height)),
				Difficulty: evmInt256.New(0),
				GasLimit:   blockGasLimit,
				Hash:       common.BytesDataToEVMIntHash(blk.BlankSeal.Hash),
			},
			Contract:    contract,
			Transaction: evmTransaction,
			Message: environment.Message{
				Caller: caller,
				Value:  evmInt256.FromDecimalString(tx.Value),
				Data:   tx.Data,
			},
		},
	}), &contract, nil
}

func (l Ledger) processEVMBalanceCache(cache storage.BalanceCache, resultCache txResultCache, newState *[]StateData) {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	balanceToChange := *newState
	for _, k := range keys {
		addr := cache[k].Address.Bytes()
		val := cache[k].Balance.Int

		var localBalance *big.Int
		var err error

		if resultCache[string(addr)] != nil {
			localBalance = resultCache[string(addr)].val
		} else {
			localBalance, err = l.BalanceOf(addr)
			if err != nil {
				continue
			}

			resultCache[string(addr)] = &txResultCacheData{
				val: localBalance,
			}
		}

		orgBalance := localBalance.Bytes()

		localBalance.Add(localBalance, val)
		balanceToChange = append(balanceToChange, StateData{
			Key:    BuildKey(StoragePrefixes.Balance, addr),
			NewVal: localBalance.Bytes(),
			OrgVal: orgBalance,
		})
	}

	*newState = balanceToChange
}

func (l Ledger) processEVMNamedStateCache(ns string, cache storage.Cache, org storage.Cache, newState *[]StateData) {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	state := *newState
	for _, k := range keys {
		cacheData := cache[k]
		orgData := org[k]

		var orgVal []byte
		if orgData != nil {
			orgVal = orgData.Bytes()
		}

		state = append(state, StateData{
			Key:    BuildKey(StoragePrefixes.ContractData, []byte(ns), []byte(k)),
			NewVal: cacheData.Bytes(),
			OrgVal: orgVal,
		})
	}

	*newState = state
}

func (l Ledger) processEVMStateCache(cache storage.CacheUnderNamespace, org storage.CacheUnderNamespace, newState *[]StateData) {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		l.processEVMNamedStateCache(k, cache[k], org[k], newState)
	}
}

func (l Ledger) processEVMLogData(ns string, logList []storage.Log, newState *[]StateData) {
	state := *newState

	for _, contractLog := range logList {
		logKey := fmt.Sprintf("%s-%x", ns, contractLog.Context.Transaction.TxHash)

		topicsCnt := byte(len(contractLog.Topics))

		logInfo := [][]byte{{topicsCnt}}
		logInfo = append(logInfo, contractLog.Topics...)
		logInfo = append(logInfo, contractLog.Data)

		logBytes := bytes.Join(logInfo, nil)
		state = append(state, StateData{
			Key:    BuildKey(StoragePrefixes.ContractLog, []byte(logKey)),
			NewVal: logBytes,
		})
	}

	*newState = state
}

func (l Ledger) processEVMLogCache(cache storage.LogCache, newState *[]StateData) {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		l.processEVMLogData(k, cache[k], newState)
	}
}

func (l Ledger) processEVMDestructs(cache storage.Cache, newState *[]StateData) {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	state := *newState

	for _, k := range keys {
		state = append(state, StateData{
			Key:    BuildKey(StoragePrefixes.ContractDestructs, []byte(k)),
			NewVal: []byte{1},
		})
	}

	*newState = state
}

func (l Ledger) newStateFromEVMResult(evmRet SealEVM.ExecuteResult, cache txResultCache) []StateData {
	evmCache := evmRet.StorageCache
	var newState []StateData

	l.processEVMBalanceCache(evmCache.Balance, cache, &newState)
	l.processEVMStateCache(evmCache.CachedData, evmCache.OriginalData, &newState)
	l.processEVMLogCache(evmCache.Logs, &newState)
	l.processEVMDestructs(evmCache.Destructs, &newState)

	return newState
}
