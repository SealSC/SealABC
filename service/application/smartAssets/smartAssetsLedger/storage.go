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
	"SealABC/storage/db/dbInterface/kvDatabase"
	"encoding/json"
)

const commonKeyLength = 32

type prefixEl struct {
	enum.Element
}

func (p prefixEl) buildKey(key []byte) []byte {
	return append([]byte(p.String()), key...)
}

var StoragePrefixes struct {
	Assets       prefixEl
	Transaction  prefixEl
	ContractData prefixEl
	Balance      prefixEl
}

func (l Ledger) getAssetsByHash(hash []byte) (assets *BaseAssets, exists bool, err error) {
	key := StoragePrefixes.Assets.buildKey(hash)
	assetsJson, err := l.Storage.Get(key)
	if err != nil {
		return
	}

	exists = assetsJson.Exists
	if exists {
		assets = &BaseAssets{}
		err = json.Unmarshal(assetsJson.Data, assets)
	}

	return
}

func (l Ledger) storeAssets(assets BaseAssets) error {
	assetsJson, err := json.Marshal(assets)
	if err != nil {
		return err
	}

	key := StoragePrefixes.Assets.buildKey(assets.MetaSeal.Hash)

	err = l.Storage.Put(kvDatabase.KVItem{
		Key:    key,
		Data:   assetsJson,
		Exists: true,
	})

	return err
}

func (l Ledger) getTxFromStorage(hash []byte) (tx *Transaction, exists bool, err error) {
	key := StoragePrefixes.Transaction.buildKey(hash)

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

//status data must start with 32 byte (commonKeyLength) store key
func (l Ledger) storeStatus(prefix prefixEl, statusList [][]byte) error {
	statusCount := len(statusList)
	var kvList = make([]kvDatabase.KVItem, statusCount, statusCount)
	for i, status := range statusList {
		txHash := status[:commonKeyLength]
		key := prefix.buildKey(txHash)
		kvList[i] = kvDatabase.KVItem{
			Key:    key,
			Data:   status[commonKeyLength:],
			Exists: true,
		}
	}
	
	return l.Storage.BatchPut(kvList)
}
