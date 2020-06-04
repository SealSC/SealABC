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
	"SealABC/metadata/seal"
	"SealABC/storage/db/dbInterface/kvDatabase"
	"encoding/json"
	"errors"
	"math/big"
)

type BaseAssetsData struct {
	Name        string
	Symbol      string
	Supply      string
	Precision   byte
	Increasable bool
	Owner       string
}

type BaseAssets struct {
	BaseAssetsData

	IssuedSeal seal.Entity
	MetaSeal   seal.Entity
}

func (b BaseAssets) getHash() []byte {
	return b.MetaSeal.Hash
}

func (l Ledger) BalanceOf(address []byte) (balance *big.Int, err error) {
	_, exists, err := l.getSystemAssets()
	if err != nil {
		return
	}

	if !exists {
		err = errors.New("no such assets")
		return
	}

	balanceKey := StoragePrefixes.Balance.BuildKey(address)
	bKV, err := l.Storage.Get(balanceKey)
	if err != nil {
		return
	}

	if !bKV.Exists {
		return big.NewInt(0), nil
	}

	balance = big.NewInt(0).SetBytes(bKV.Data)
	return
}

func (l Ledger) getSystemAssets() (assets *BaseAssets, exits bool, err error) {
	key := StoragePrefixes.SystemAssets.BuildKey(nil)

	assetsJson, err := l.Storage.Get(key)
	if err != nil {
		return
	}

	exits = assetsJson.Exists

	if exits {
		assets = &BaseAssets{}
		err = json.Unmarshal(assetsJson.Data, assets)
	}

	return
}

func (l Ledger) getAssetsByHash(hash []byte) (assets *BaseAssets, exists bool, err error) {
	key := StoragePrefixes.Assets.BuildKey(hash)
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

func (l Ledger) storeSystemAssets(assets BaseAssets) error {
	assetsJson, err := json.Marshal(assets)
	if err != nil {
		return err
	}

	key := StoragePrefixes.SystemAssets.BuildKey(nil)

	err = l.Storage.Put(kvDatabase.KVItem{
		Key:    key,
		Data:   assetsJson,
		Exists: true,
	})

	return err
}

func (l Ledger) storeAssets(assets BaseAssets) error {
	assetsJson, err := json.Marshal(assets)
	if err != nil {
		return err
	}

	key := StoragePrefixes.Assets.BuildKey(assets.MetaSeal.Hash)

	err = l.Storage.Put(kvDatabase.KVItem{
		Key:    key,
		Data:   assetsJson,
		Exists: true,
	})

	return err
}
