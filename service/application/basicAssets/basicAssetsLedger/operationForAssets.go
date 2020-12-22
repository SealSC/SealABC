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
    "github.com/SealSC/SealABC/common"
    "bytes"
    "errors"
    "time"
)

func (l *Ledger) verifyIncreaseSupply(tx Transaction) (ret interface{}, err error) {
    assets := tx.Assets
    err = assets.verify(l.CryptoTools)
    if err != nil {
        return
    }

    localAssets, err := l.getLocalAssets(assets)
    if err != nil {
        err = errors.New("assets not exist: " + err.Error())
        return
    }

    if !l.assetsExists(assets) {
        err = errors.New("can't increase assets supply due to no such assets")
        return
    }

    if !bytes.Equal(localAssets.IssuedSeal.SignerPublicKey, assets.IssuedSeal.SignerPublicKey) {
        err = errors.New("invalid owner")
        return
    }

    if assets.Supply <= localAssets.Supply {
        err = errors.New("invalid new supply")
        return
    }

    return
}

func (l *Ledger) confirmIncreaseSupply(tx Transaction) (ret interface{}, err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    ret, err = l.verifyIncreaseSupply(tx)
    if err != nil {
        return
    }

    assets := tx.Assets
    err = l.updateAssets(assets)
    return
}

func (l *Ledger) verifyIssueAssets(tx Transaction) (ret interface{}, err error) {
    if tx.TxType != TransactionTypes.IssueAssets.String() {
        err = errors.New("invalid transaction type")
        return
    }

    assets := tx.Assets
    if l.assetsExists(assets) {
        err = errors.New("can't push issue asset transaction, assets already exists")
        return
    }

    err = assets.verify(l.CryptoTools)
    if err != nil {
        return
    }

    return
}

func (l *Ledger) confirmIssueAssets(tx Transaction) (ret interface{}, err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    if tx.TxType != TransactionTypes.IssueAssets.String() {
        err = errors.New("invalid transaction type")
        return
    }

    assets := tx.Assets
    err = assets.verify(l.CryptoTools)
    if err != nil {
        return
    }

    assets.DateTime = time.Now().Format(common.BASIC_TIME_FORMAT)
    err = l.storeAssets(assets)
    if err != nil {
        return
    }

    if uint32(AssetsTypes.Copyright.Int()) == assets.Type {
        err = l.storeCopyright(assets, tx.Seal.SignerPublicKey)
        if err != nil {
            return
        }
    }

    ret, err = l.saveUnspentInsideIssueAssetsTransaction(tx)

    return
}
