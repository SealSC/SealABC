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
	"SealABC/storage/db/dbInterface/kvDatabase"
	"errors"
	"math/big"
)

var bigZero = big.NewInt(0)

func (l Ledger) verifyTransfer(tx Transaction, cacheData interface{}) (interface{}, error) {
	if tx.Type != TxType.Transfer.String() {
		return nil, errors.New("invalid transaction type")
	}

	_, err := tx.DataSeal.Verify(tx.getData(), l.CryptoTools.HashCalculator)
	if err != nil {
		return nil, err
	}

	var balance *big.Int
	if cacheData != nil {
		balance = cacheData.(*big.Int)
	} else {
		balance, err = l.balanceOf(tx.From, l.genesisAssets.getHash())
		if err != nil {
			return nil, err
		}
	}


	if balance.Cmp(bigZero) <= 0 {
		return nil, errors.New("insufficient balance")
	}

	amount, valid := big.NewInt(0).SetString(string(tx.Value), 10)
	if !valid {
		return nil, errors.New("invalid transfer value")
	}

	if amount.Sign() < 0 {
		return nil, errors.New("negative transfer value")
	} else if amount.Sign() == 0 {
		return balance, nil
	}

	if balance.Cmp(amount) < 0 {
		return nil, errors.New("insufficient balance")
	}

	balance.Sub(balance, amount)
	return balance, nil
}

func (l Ledger) PreTransfer(tx Transaction, preResult map[string] *big.Int) (result map[string] *big.Int, err error) {
	if preResult == nil {
		preResult = map[string] *big.Int{}
	}

	addressStr := string(tx.From)
	balance, err := l.verifyTransfer(tx, preResult[addressStr])
	if err != nil {
		return preResult, err
	}

	preResult[addressStr] = balance.(*big.Int)
	return preResult, err
}

func (l Ledger) SaveTransferResult(txList []Transaction) (err error) {

	var statusKVList []kvDatabase.KVItem

	for _, tx := range txList {
		for _, s := range tx.TransactionResult.NewStatus {
			statusKVList = append(statusKVList, kvDatabase.KVItem {
				Key:    s.Key,
				Data:   s.Val,
				Exists: true,
			})
		}
	}

	err = l.Storage.BatchPut(statusKVList)
	return
}
