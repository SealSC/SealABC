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
    "SealABC/crypto"
    "SealABC/crypto/hashes/sha3"
    "SealABC/crypto/signers/ed25519"
    "SealABC/dataStructure/enum"
    "SealABC/metadata/blockchainRequest"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "sync"
)

var StoragePrefixes struct{
    Assets                      enum.Element
    Unspent                     enum.Element
    Transaction                 enum.Element
    TransactionWithBlockInfo    enum.Element

    Balance                     enum.Element
}

type txValidator func(tx Transaction) (ret interface{}, err error)
type txActuator func(tx Transaction) (ret interface{}, err error)
type ledgerQuery func(param []string) (ret interface{}, err error)

type Ledger struct {
    operateLock sync.RWMutex
    poolLock    sync.Mutex

    txPool          map[string] blockchainRequest.Entity
    utxoPool        map[string] bool
    txValidators    map[string] txValidator
    txActuators     map[string] txActuator
    ledgerQueries   map[string] ledgerQuery

    CryptoTools crypto.Tools
    Storage     kvDatabase.IDriver
}

func Load()  {
    enum.SimpleBuild(&TransactionTypes)
    enum.SimpleBuild(&StoragePrefixes)
    enum.SimpleBuild(&QueryTypes)
}

func NewLedger(storage kvDatabase.IDriver) (ledger *Ledger) {
    ledger = &Ledger{}

    ledger.Storage = storage

    ledger.CryptoTools = crypto.Tools{
        HashCalculator:  sha3.Sha256,
        SignerGenerator: ed25519.SignerGenerator,
    }

    ledger.txPool = map[string] blockchainRequest.Entity{}
    ledger.utxoPool = map[string] bool{}

    ledger.txValidators = map[string] txValidator {
        TransactionTypes.IssueAssets.String(): ledger.verifyIssueAssets,
        TransactionTypes.IncreaseSupply.String(): ledger.verifyIncreaseSupply,
        TransactionTypes.Transfer.String(): ledger.verifyTransfer,
    }

    ledger.txActuators = map[string] txActuator {
        TransactionTypes.IssueAssets.String(): ledger.confirmIssueAssets,
        TransactionTypes.IncreaseSupply.String(): ledger.confirmIncreaseSupply,
        TransactionTypes.Transfer.String(): ledger.confirmTransfer,
    }

    ledger.ledgerQueries = map[string] ledgerQuery {
        QueryTypes.Assets.String(): ledger.queryAssets,
        QueryTypes.AllAssets.String(): ledger.queryAllAssets,
        QueryTypes.UnspentList.String(): ledger.queryUnspent,
        QueryTypes.Transaction.String(): ledger.queryTransaction,
    }

    return
}
