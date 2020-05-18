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
    "SealABC/metadata/seal"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "bytes"
    "encoding/json"
    "errors"
)

type AssetsData struct {
    Name        string
    Symbol      string
    Supply      uint64 `json:",string"`
    Increasable bool
}

type Assets struct {
    AssetsData

    IssuedSeal seal.Entity
    MetaSeal   seal.Entity
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

    _, err = a.IssuedSeal.Verify(fullBytes, tools)
    if err != nil {
        err = errors.New("invalid full assets data signature: " + err.Error())
        return
    }

    _, err = a.MetaSeal.Verify(withoutSupplyBytes, tools)
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

    _, err := l.Storage.Get(key)

    return err == nil
}

func (l *Ledger) newAssets(assets Assets) (err error) {
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

    err = json.Unmarshal(kv.Data, &a)
    return
}
