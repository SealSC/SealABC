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
    "SealABC/metadata/blockchainRequest"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "encoding/json"
    "errors"
    "time"
)

func (l *Ledger) buildTransactionKey(txHash []byte) (key [] byte) {
    //prefix + sender + transaction hash
    key = []byte(StoragePrefixes.Transaction.String())
    key = append(key, txHash...)
    return
}

func (l *Ledger) verifyTransaction(tx Transaction) (err error) {
    validator, exists := l.txValidators[tx.TxType]
    if !exists {
        err = errors.New("no validator for this transaction: " + tx.TxType)
        return
    }

    _, err = validator(tx)
    if err != nil {
        log.Log.Error("invalid transaction: ", tx)
        return
    }

    return
}

func (l *Ledger) getLocalTransaction(txHash []byte) (tx TransactionWithBlockInfo, err error) {
    key := l.buildTransactionKey(txHash)

    kv, err := l.Storage.Get(key)
    if err != nil {
        return
    }

    err = json.Unmarshal(kv.Data, &tx)
    return
}

func (l *Ledger) PushTransaction(req blockchainRequest.Entity) (err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    tx := Transaction{}
    err = json.Unmarshal(req.Data, &tx)
    if err != nil {
        return
    }

    err = l.verifyTransaction(tx)
    if err != nil {
        return
    }

    l.poolLock.Lock()
    defer l.poolLock.Unlock()

    tx.CreateTime = time.Now().Unix()
    l.txPool[tx.HashString()] = req
    return
}

func (l *Ledger) GetTransactionsFromPool() (txList []blockchainRequest.Entity, count uint32) {
    l.poolLock.Lock()
    defer l.poolLock.Unlock()

    count = 0
    for _, tx := range l.txPool {
        count += 1
        txList = append(txList, tx)
    }

    return
}

func (l *Ledger) RemoveTransactionFromPool(txKey string) {
    l.poolLock.Lock()
    defer l.poolLock.Unlock()
    delete(l.txPool, txKey)

    return
}

//wrap verify transaction method to solve interlock
func (l *Ledger) VerifyTransaction(tx Transaction) (err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    err = l.verifyTransaction(tx)
    if err != nil {
        return
    }

    return
}

func (l *Ledger) ExecuteTransaction(tx Transaction) (ret interface{}, err error) {
    handle, exists := l.txActuators[tx.TxType]
    if !exists {
        err = errors.New("no actuator for this transaction: " + tx.TxType)
        return
    }

    ret, err = handle(tx)
    if err != nil {
        log.Log.Error("execute transaction failed: ", tx)
        return
    }

    return
}

func (l *Ledger) GetOriginalTransactionWithBlockInfo(hash []byte) (tx TransactionWithBlockInfo, err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    key := l.buildTransactionKey(hash)
    kv, err := l.Storage.Get(key)

    if err != nil {
        return
    }

    err = json.Unmarshal(kv.Data, &tx)
    return
}

func (l *Ledger) SaveTransactionWithBlockInfo(tx Transaction, reqHash []byte, blockHeight uint64, actIndex uint32) (err error) {
    l.operateLock.Lock()
    defer l.operateLock.Unlock()

    key := l.buildTransactionKey(tx.Seal.Hash)

    txWithBlkInfo := TransactionWithBlockInfo {
        Transaction:tx,
    }

    txWithBlkInfo.BlockInfo.RequestHash = reqHash
    txWithBlkInfo.BlockInfo.BlockHeight = blockHeight
    txWithBlkInfo.BlockInfo.ActionIndex = actIndex

    txBytes, err := json.Marshal(txWithBlkInfo)

    if err != nil {
        return
    }

    err = l.Storage.Put(kvDatabase.KVItem{
        Key:key,
        Data: txBytes,
    })

    return
}
