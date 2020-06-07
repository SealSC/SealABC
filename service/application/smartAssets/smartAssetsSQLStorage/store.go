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

package smartAssetsSQLStorage

import (
	"SealABC/log"
	"SealABC/metadata/block"
	"SealABC/service/application/smartAssets/smartAssetsLedger"
	"SealABC/service/application/smartAssets/smartAssetsSQLTables"
	"bytes"
	"math/big"
)

func (s Storage) getClassifiedTableRows(txType string) smartAssetsSQLTables.ISQLRows {
	switch txType {
	case smartAssetsLedger.TxType.Transfer.String():
		rows := smartAssetsSQLTables.Transfer.NewRows().(smartAssetsSQLTables.TransferRows)
		return &rows

	case smartAssetsLedger.TxType.ContractCall.String():
		rows := smartAssetsSQLTables.ContractCall.NewRows().(smartAssetsSQLTables.ContractCallRows)
		return &rows

	case smartAssetsLedger.TxType.CreateContract.String():
		rows := smartAssetsSQLTables.Contract.NewRows().(smartAssetsSQLTables.ContractRows)
		return &rows
	}

	return nil
}

func (s Storage) isNewBalance(key []byte) bool {
	balancePrefixKey := smartAssetsLedger.BuildKey(smartAssetsLedger.StoragePrefixes.Balance, nil)
	return bytes.Equal(balancePrefixKey, key[:len(balancePrefixKey)])
}

func (s Storage) StoreSystemIssueBalance(balance *big.Int, owner string) error {
	addressListRows := smartAssetsSQLTables.AddressList.NewRows().(smartAssetsSQLTables.AddressListRows)
	addressListRows.InsertSystemIssueBalance(balance, owner)

	_, err := s.Driver.Insert(&addressListRows, true)
	return err
}

func (s Storage) StoreTransaction(tx smartAssetsLedger.Transaction, blk block.Entity) (err error) {
	txRows := smartAssetsSQLTables.Transaction.NewRows().(smartAssetsSQLTables.TransactionRows)

	txRows.Insert(tx, blk)
	_, err = s.Driver.Insert(&txRows, true)
	if err != nil {
		log.Log.Error("insert transaction to sql database failed: ", err.Error())
	}
	
	classifiedRows := s.getClassifiedTableRows(tx.Type)
	if classifiedRows == nil {
		return
	}

	classifiedRows.Insert(tx, blk)
	_, err = s.Driver.Insert(classifiedRows, true)
	if err != nil {
		log.Log.Error("insert classified rows [" + tx.Type + "] failed: " + err.Error())
	}


	addressListRows := smartAssetsSQLTables.AddressList.NewRows().(smartAssetsSQLTables.AddressListRows)
	for _, v := range tx.TransactionResult.NewState {
		if s.isNewBalance(v.Key) {
			balance := big.NewInt(0).SetBytes(v.Val)
			addressListRows.Insert(v.Key, balance, blk)
		}
	}

	_, err = s.Driver.Replace(&addressListRows)
	if err != nil {
		log.Log.Error("update balance rows failed: " + err.Error())
	}

	return
}
