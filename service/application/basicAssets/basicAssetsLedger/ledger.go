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
    "github.com/SealSC/SealABC/crypto"
    "github.com/SealSC/SealABC/crypto/hashes/sha3"
    "github.com/SealSC/SealABC/crypto/signers/ed25519"
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/metadata/blockchainRequest"
    "github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
    "sync"
)

var StoragePrefixes struct{
    Assets                   enum.Element
    Copyright                enum.Element
    Unspent                  enum.Element
    Transactions             enum.Element
    TransactionWithBlockInfo enum.Element

    Balance     enum.Element
    SellingList enum.Element
}

type txValidator func(tx Transaction) (ret interface{}, err error)
type txActuator func(tx Transaction) (ret interface{}, err error)
type ledgerQuery func(param []string) (ret interface{}, err error)

type Ledger struct {
    operateLock sync.RWMutex
    poolLock    sync.Mutex

    txPool         map[string] blockchainRequest.Entity
    memUTXORecord  map[string] bool
    execUTXORecord map[string] bool
    txValidators   map[string] txValidator
    txActuators     map[string] txActuator
    ledgerQueries   map[string] ledgerQuery

    CryptoTools crypto.Tools
    Storage     kvDatabase.IDriver
}

func Load()  {
    enum.SimpleBuild(&TransactionTypes)
    enum.SimpleBuild(&StoragePrefixes)
    enum.SimpleBuild(&QueryTypes)
    enum.SimpleBuild(&AssetsTypes)
}

func NewLedger(storage kvDatabase.IDriver) (ledger *Ledger) {
    ledger = &Ledger{}

    ledger.Storage = storage

    ledger.CryptoTools = crypto.Tools{
        HashCalculator:  sha3.Sha256,
        SignerGenerator: ed25519.SignerGenerator,
    }

    ledger.txPool = map[string] blockchainRequest.Entity{}
    ledger.memUTXORecord = map[string] bool{}
    ledger.execUTXORecord = map[string] bool{}

    ledger.txValidators = map[string] txValidator {
        TransactionTypes.IssueAssets.String(): ledger.verifyIssueAssets,
        TransactionTypes.IncreaseSupply.String(): ledger.verifyIncreaseSupply,
        TransactionTypes.Transfer.String(): ledger.verifyTransfer,

        TransactionTypes.StartSelling.String(): ledger.verifyStartSelling,
        TransactionTypes.StopSelling.String(): ledger.verifyStopSelling,
        TransactionTypes.BuyAssets.String(): ledger.verifyBuyAssets,
    }

    ledger.txActuators = map[string] txActuator {
        TransactionTypes.IssueAssets.String(): ledger.confirmIssueAssets,
        TransactionTypes.IncreaseSupply.String(): ledger.confirmIncreaseSupply,
        TransactionTypes.Transfer.String(): ledger.confirmTransfer,

        TransactionTypes.StartSelling.String(): ledger.confirmStartSelling,
        TransactionTypes.StopSelling.String(): ledger.confirmStopSelling,
        TransactionTypes.BuyAssets.String(): ledger.confirmBuyAssets,
    }

    ledger.ledgerQueries = map[string] ledgerQuery {
        QueryTypes.Assets.String(): ledger.queryAssets,
        QueryTypes.AllAssets.String(): ledger.queryAllAssets,
        QueryTypes.UnspentList.String(): ledger.queryUnspent,
        QueryTypes.Transaction.String(): ledger.queryTransaction,
        QueryTypes.SellingList.String(): ledger.querySellingList,
        QueryTypes.Copyright.String(): ledger.queryCopyright,
    }

    return
}
