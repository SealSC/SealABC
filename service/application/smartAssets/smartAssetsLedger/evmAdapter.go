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
	"SealABC/metadata/block"
	"SealEVM"
	"SealEVM/common"
	"SealEVM/environment"
	"SealEVM/evmInt256"
	"SealEVM/storage"
	"bytes"
	"fmt"
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

func (l Ledger)newEVM(tx Transaction, callback SealEVM.EVMResultCallback,
	blk block.Entity, blockGasLimit *evmInt256.Int) (*SealEVM.EVM, environment.Contract) {

	evmTransaction := environment.Transaction{
		Origin:   common.BytesDataToEVMIntHash(tx.DataSeal.Hash),
		GasPrice: constTransactionGasPrice(),
		GasLimit: constTransactionGasLimit(),
	}

	hashByte := l.CryptoTools.HashCalculator.Sum(tx.Data)
	contractHash := common.BytesDataToEVMIntHash(hashByte)
	caller := common.BytesDataToEVMIntHash(tx.DataSeal.Hash)
	var contractAddress *evmInt256.Int
	if len(tx.To) == 0 {
		contractAddress = l.storageForEVM.CreateAddress(caller, evmTransaction)
	} else {
		contractAddress = common.BytesDataToEVMIntHash(tx.To)
	}

	contract := environment.Contract{
		Namespace: contractAddress,
		Code:      tx.Data,
		Hash:      contractHash,
	}

	return SealEVM.New(SealEVM.EVMParam{
		MaxStackDepth:  defaultStackDepth,
		ExternalStore:  &l.storageForEVM,
		ResultCallback: callback,
		Context:        &environment.Context{
			Block:       environment.Block{
				Coinbase:   common.BytesDataToEVMIntHash(blk.BlankSeal.SignerPublicKey),
				Timestamp:  evmInt256.New(int64(blk.Header.Timestamp)),
				Number:     evmInt256.New(int64(blk.Header.Height)),
				Difficulty: evmInt256.New(0),
				GasLimit:   blockGasLimit,
				Hash:       common.BytesDataToEVMIntHash(blk.BlankSeal.Hash),
			},
			Contract:    contract,
			Transaction: evmTransaction,
			Message:     environment.Message {
				Caller: caller,
				Value:  evmInt256.FromDecimalString(tx.Value),
				Data:   nil,
			},
		},
	}), contract
}

func getSortedKeys(src interface{}) []string {
	srcAsMap, ok := src.(map[string] interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(srcAsMap))
	for k := range srcAsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (l Ledger) processEVMBalanceCache(cache storage.BalanceCache, resultCache txResultCache, newState *[]StateData) {
	keys := getSortedKeys(cache)

	balanceToChange := *newState
	assetsHash := l.genesisAssets.getHash()
	for _, k := range keys {
		addr := cache[k].Address.Bytes()
		val := cache[k].Balance.Int

		var localBalance *big.Int
		var err error

		if resultCache[string(addr)] != nil {
			localBalance = resultCache[string(addr)].val
		} else {
			localBalance, err = l.balanceOf(addr, assetsHash)
			if err != nil {
				continue
			}


			resultCache[string(addr)] = &txResultCacheData{
				val: localBalance,
			}
		}

		localBalance.Add(localBalance, val)
		balanceToChange = append(balanceToChange, StateData{
			Key: StoragePrefixes.Balance.buildKey(addr),
			Val: localBalance.Bytes(),
		})
	}

	*newState = balanceToChange
}

func (l Ledger) processEVMNamedStateCache(ns string, cache storage.Cache, newState *[]StateData) {
	state := *newState
	keys := getSortedKeys(cache)

	for _, k := range keys {
		cacheData := cache[k]
		state = append(state, StateData{
			Key: StoragePrefixes.ContractData.buildKey([]byte(ns), []byte(k)),
			Val: cacheData.Bytes(),
		})

	}

	*newState = state
}

func (l Ledger) processEVMStateCache(cache storage.CacheUnderNamespace, newState *[]StateData) {
	contractKeys := getSortedKeys(cache)

	for _, k := range contractKeys {
		l.processEVMNamedStateCache(k, cache[k], newState)
	}
}

func (l Ledger) processEVMLogData(ns string, logList []storage.Log, newState *[]StateData) {
	state := *newState

	for _, log := range logList {
		logKey := fmt.Sprintf("%s-%x", ns, log.Context.Transaction.TxHash)

		topicsCnt := byte(len(log.Topics))

		logInfo := [][]byte{{topicsCnt}}
		logInfo = append(logInfo, log.Topics...)
		logInfo = append(logInfo, log.Data)

		logBytes := bytes.Join(logInfo, nil)
		state = append(state, StateData{
			Key: StoragePrefixes.ContractLog.buildKey([]byte(logKey)),
			Val: logBytes,
		})
	}

	*newState = state
}

func (l Ledger) processEVMLogCache(cache storage.LogCache, newState *[]StateData) {
	keys := getSortedKeys(cache)

	for _, k := range keys {
		l.processEVMLogData(k, cache[k], newState)
	}
}

func (l Ledger) processEVMDestructs(cache storage.Cache, newState *[]StateData) {
	keys := getSortedKeys(cache)
	state := *newState

	for _, k := range keys {
		state = append(state, StateData{
			Key: StoragePrefixes.ContractDestructs.buildKey([]byte(k)),
			Val: []byte{1},
		})
	}

	*newState = state
}

func (l Ledger) newStateFromEVMResult(evmRet SealEVM.ExecuteResult, cache txResultCache) []StateData {
	evmCache := evmRet.StorageCache
	var newState []StateData

	l.processEVMBalanceCache(evmCache.Balance, cache, &newState)
	l.processEVMStateCache(evmCache.CachedData, &newState)
	l.processEVMLogCache(evmCache.Logs, &newState)
	l.processEVMDestructs(evmCache.Destructs, &newState)

	return newState
}
