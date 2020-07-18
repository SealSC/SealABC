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
    "SealABC/log"
    "encoding/base64"
    "encoding/json"
    "errors"
)

func (l *Ledger)DoQuery(req QueryRequest) (data interface{}, err error) {
    l.operateLock.RLock()
    defer l.operateLock.RUnlock()

    queryHandle, exists := l.ledgerQueries[req.QueryType]
    if !exists {
        err = errors.New("no query action named " + req.QueryType)
        return
    }

    return queryHandle(req.Parameter)
}

func (l *Ledger) queryAssets(p []string) (result interface{}, err error) {
    var hashBytes []byte
    err = json.Unmarshal([]byte(p[0]), &hashBytes)
    if err != nil {
        return
    }

    result, err = l.localAssetsFromHash(hashBytes)
    return
}

func (l *Ledger) queryAllAssets(_ []string) (result interface{}, err error) {
    prefix := []byte(StoragePrefixes.Assets.String())

    kvList := l.Storage.Traversal(prefix)
    assetsList := AssetsList{}
    for _, kv := range kvList {
        assets := Assets{}
        _ = json.Unmarshal(kv.Data, &assets)
        assetsList.List = append(assetsList.List, assets)
    }

    result = assetsList
    return
}

func (l *Ledger) queryUnspent(p []string) (result interface{}, err error) {
    queryParam := UnspentQueryParameter{}
    err = json.Unmarshal([]byte(p[0]), &queryParam)
    if err != nil {
        log.Log.Error("bad request: ", err.Error())
        return
    }

    prefix := l.buildUnspentQueryPrefix(queryParam.Address, queryParam.Assets)
    kvList := l.Storage.Traversal(prefix)

    unspentList := UnspentList{}
    unspentList.List = map[string] *UnspentUnderAssets{}

    for _, kv := range kvList {
        u := Unspent{}
        err = json.Unmarshal(kv.Data, &u)
        if err != nil {
            log.Log.Error("unmarshal u from storage failed: ", err.Error())
            continue
        }

        assetsHash := u.AssetsHash
        assetsHashString := base64.StdEncoding.EncodeToString(assetsHash)
        _, exist := unspentList.List[assetsHashString]

        if !exist {
            assets, err := l.localAssetsFromHash(assetsHash)
            if err != nil {
                log.Log.Error("get local assets ", assetsHashString, " failed: ", err.Error())
                continue
            }
            unspentList.List[assetsHashString] = &UnspentUnderAssets{
                Assets: assets,
            }
        }

        unspentList.List[assetsHashString].UnspentList = append(unspentList.List[assetsHashString].UnspentList, u)
    }

    result = unspentList
    return
}

func (l *Ledger) queryTransaction(p []string) (result interface{}, err error) {
    var hashBytes []byte
    err = json.Unmarshal([]byte(p[0]), &hashBytes)
    if err != nil {
        return
    }

    result, err = l.getLocalTransaction(hashBytes)
    return
}

func (l *Ledger) querySellingList(_ []string) (result interface{}, err error) {
    list := l.Storage.Traversal([]byte(StoragePrefixes.SellingList.String()))

    var sellingList []SellingData

    for _, data := range list {
        sellingData := SellingData{}
        jsonErr := json.Unmarshal(data.Data, &sellingData)
        if jsonErr != nil {
            continue
        }
        sellingList = append(sellingList, sellingData)
    }
    return sellingList, nil
}

func (l *Ledger) queryCopyright(_ []string) (result interface{}, err error) {
    list := l.Storage.Traversal([]byte(StoragePrefixes.Copyright.String()))

    var copyrightList []Copyright

    for _, data := range list {
        cr := Copyright{}
        jsonErr := json.Unmarshal(data.Data, &cr)
        if jsonErr != nil {
            continue
        }
        copyrightList = append(copyrightList, cr)
    }
    return copyrightList, nil
}
