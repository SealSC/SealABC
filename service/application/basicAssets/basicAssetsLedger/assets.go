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

package basicAssetsLedger

import (
    "SealABC/common/utility/serializer/structSerializer"
    "SealABC/crypto"
    "SealABC/dataStructure/enum"
    "SealABC/metadata/seal"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "bytes"
    "encoding/json"
    "errors"
)

var AssetsTypes struct{
    General   enum.Element
    Copyright enum.Element
}

type AssetsData struct {
    Name        string
    Symbol      string
    Supply      uint64 `json:",string"`
    Type        uint32 `json:",string"`
    Increasable bool
    ExtraInfo   []byte
}

type Assets struct {
    AssetsData

    DateTime   string
    IssuedSeal seal.Entity
    MetaSeal   seal.Entity
}

type Copyright struct {
    Assets
    Owner  []byte
}

func (a Assets) getUniqueHash() (hash []byte) {
    return a.MetaSeal.Hash
}

func (a Assets) verify(tools crypto.Tools) (err error) {
    if !bytes.Equal(a.IssuedSeal.SignerPublicKey, a.MetaSeal.SignerPublicKey) {
        err = errors.New("with and without supply's signer are not equal")
        return
    }

    fullBytes, err := structSerializer.ToMFBytes(a.AssetsData)
    if err != nil {
        return
    }

    a.Supply = 0
    withoutSupplyBytes, _ := structSerializer.ToMFBytes(a.AssetsData)

    _, err = a.IssuedSeal.Verify(fullBytes, tools.HashCalculator)
    if err != nil {
        err = errors.New("invalid full assets data signature: " + err.Error())
        return
    }

    _, err = a.MetaSeal.Verify(withoutSupplyBytes, tools.HashCalculator)
    if err != nil {
        err = errors.New("invalid assets without supply data signature: " + err.Error())
        return
    }

    return
}

func (l *Ledger) buildAssetsKey(assets Assets) (key []byte) {
    //prefix + without supply hash
    key = []byte(StoragePrefixes.Assets.String())
    key = append(key, assets.getUniqueHash()...)
    return
}

func (l *Ledger) assetsExists(assets Assets) bool {
    key := l.buildAssetsKey(assets)

    assetsData, _ := l.Storage.Get(key)
    exists := assetsData.Exists
    return exists
}

func (l *Ledger) storeAssets(assets Assets) (err error) {
    if l.assetsExists(assets) {
        err = errors.New("can't issue new assets due to the assets already exists")
        return
    }

    key := l.buildAssetsKey(assets)
    assetsBytes, _ := json.Marshal(assets)

    err = l.Storage.Put(kvDatabase.KVItem{
        Key: key,
        Data: assetsBytes,
    })

    return
}

func (l *Ledger) updateAssets(assets Assets) (err error) {

    if !l.assetsExists(assets) {
        err = errors.New("assets not exists")
        return
    }

    key := l.buildAssetsKey(assets)
    assetsBytes, _ := json.Marshal(assets)

    err = l.Storage.Put(kvDatabase.KVItem{
        Key: key,
        Data: assetsBytes,
    })

    return
}

func (l *Ledger) localAssetsFromHash(hash []byte) (assets Assets, err error) {
    assets.MetaSeal.Hash = hash
    assets, err = l.getLocalAssets(assets)

    return
}

func (l *Ledger) getLocalAssets(assets Assets) (a Assets, err error){
    assetsKey := l.buildAssetsKey(assets)

    kv, err := l.Storage.Get(assetsKey)
    if err != nil {
        return
    }

    if !kv.Exists {
        err = errors.New("no such assets")
        return
    }

    err = json.Unmarshal(kv.Data, &a)
    return
}
