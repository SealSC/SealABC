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
	"encoding/json"
	"github.com/SealSC/SealABC/metadata/seal"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
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

func (b *BaseAssets) getHash() []byte {
	return b.MetaSeal.Hash
}

func (l *Ledger) getSystemAssets() (assets *BaseAssets, exits bool, err error) {
	key := BuildKey(StoragePrefixes.SystemAssets, nil)

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

func (l *Ledger) getAssetsByHash(hash []byte) (assets *BaseAssets, exists bool, err error) {
	key := BuildKey(StoragePrefixes.Assets, hash)
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

func (l *Ledger) storeSystemAssets(assets BaseAssets) error {
	assetsJson, err := json.Marshal(assets)
	if err != nil {
		return err
	}

	key := BuildKey(StoragePrefixes.SystemAssets, nil)

	err = l.Storage.Put(kvDatabase.KVItem{
		Key:    key,
		Data:   assetsJson,
		Exists: true,
	})

	return err
}

func (l *Ledger) storeAssets(assets BaseAssets) error {
	assetsJson, err := json.Marshal(assets)
	if err != nil {
		return err
	}

	key := BuildKey(StoragePrefixes.Assets, assets.MetaSeal.Hash)

	err = l.Storage.Put(kvDatabase.KVItem{
		Key:    key,
		Data:   assetsJson,
		Exists: true,
	})

	return err
}
