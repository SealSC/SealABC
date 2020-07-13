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

package basicAssetsInterface

import (
	"SealABC/log"
	"SealABC/service/application/basicAssets/basicAssetsLedger"
)

func (b *BasicAssetsApplication) storeAssets(tx basicAssetsLedger.TransactionWithBlockInfo, blc interface{})  {
	balance, ok := blc.(basicAssetsLedger.Balance)
	if !ok {
		log.Log.Warn("no balance")
		return
	}
	err := b.SQLStorage.StoreAssets(tx)
	if err != nil {
		log.Log.Warn("save assets result: ", err.Error())
	}

	err = b.SQLStorage.StoreBalance(tx.BlockInfo.BlockHeight, tx.CreateTime, []basicAssetsLedger.Balance{balance})
	if err != nil {
		log.Log.Warn("save assets result: ", err.Error())
	}
}

func (b *BasicAssetsApplication) storeTransfer(tx basicAssetsLedger.TransactionWithBlockInfo, list interface{})  {
	usList, ok := list.(basicAssetsLedger.UnspentListWithBalance)
	if !ok {
		log.Log.Warn("transaction has no unspent")
		return
	}

	err := b.SQLStorage.StoreUnspent(tx, usList.UnspentList)
	if err != nil {
		log.Log.Warn("save assets failed: ", err.Error())
	}

	err = b.SQLStorage.StoreBalance(tx.BlockInfo.BlockHeight, tx.CreateTime, usList.BalanceList)
	if err != nil {
		log.Log.Warn("save transfer failed: ", err.Error())
	}
}

func (b *BasicAssetsApplication) storeSelling(tx basicAssetsLedger.TransactionWithBlockInfo, execResult interface{}) {
	sellRet := execResult.(basicAssetsLedger.SellingOperationResult)

	err := b.SQLStorage.StoreBalance(tx.BlockInfo.BlockHeight, tx.CreateTime, sellRet.BalanceList)
	if err != nil {
		log.Log.Warn("save selling balance failed: ", err.Error())
	}

	err = b.SQLStorage.StoreSelling(tx, sellRet)
	if err != nil {
		log.Log.Warn("save selling data failed: ", err.Error())
	}

	return
}

