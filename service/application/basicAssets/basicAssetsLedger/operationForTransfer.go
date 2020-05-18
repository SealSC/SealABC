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
    "errors"
    "fmt"
)

func (l *Ledger) verifyTransfer(tx Transaction) (ret interface{}, err error) {
    unspentList, totalIn, err := l.getUnspentListFromTransaction(tx)

    if err != nil {
        return
    }

    //verify transfer input is equal to output
    var totalOut uint64 = 0
    for _, output := range tx.Output {
        totalOut += output.Value
    }

    if totalIn != totalOut {
        err = errors.New("input not equal output")
        return
    }

    for _, utxo := range unspentList {
        key := string(utxo.Transaction) + fmt.Sprintf("%d", utxo.OutputIndex)
        if l.utxoPool[key] {
            err = errors.New("double spent")
            break
        }
    }

    if err == nil {
        for _, utxo := range unspentList {
            key := string(utxo.Transaction) + fmt.Sprintf("%d", utxo.OutputIndex)
            l.utxoPool[key] = true
        }
    }

    ret = unspentList
    return
}

func (l *Ledger) confirmTransfer(tx Transaction) (ret interface{}, err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    //verifyTransfer transaction
    usList, err := l.verifyTransfer(tx)
    if err != nil {
        return
    }

    localAssets, _ := l.localAssetsFromHash(tx.Assets.getUniqueHash())

    //save transaction
    ret, err = l.saveUnspent(localAssets, tx, usList.([]Unspent))
    return
}

