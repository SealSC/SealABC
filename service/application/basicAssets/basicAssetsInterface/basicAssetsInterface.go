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

package basicAssetsInterface

import (
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/metadata/applicationResult"
    "github.com/SealSC/SealABC/metadata/block"
    "github.com/SealSC/SealABC/metadata/blockchainRequest"
    "github.com/SealSC/SealABC/service"
    "github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsLedger"
    "github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsSQLStorage"
    "github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
    "github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
    "github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "encoding/json"
    "errors"
)

var QueryDBType struct{
    KV  enum.Element
    SQL enum.Element
}

type BasicAssetsApplication struct {
    Ledger *basicAssetsLedger.Ledger
    SQLStorage *basicAssetsSQLStorage.Storage
}

func (b *BasicAssetsApplication) Name() (name string) {
    return "Basic Assets"
}

func (b *BasicAssetsApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
    err = b.Ledger.PushTransaction(req)
    return
}

func (b *BasicAssetsApplication) Query(req []byte) (result interface{}, err error) {
    queryReq := basicAssetsLedger.QueryRequest{}
    err = json.Unmarshal(req, &queryReq)
    if err != nil {
        return
    }

    switch queryReq.DBType {
    case QueryDBType.KV.String():
        return b.Ledger.DoQuery(queryReq)

    case QueryDBType.SQL.String():
        return b.SQLStorage.DoQuery(queryReq)

    default:
        err = errors.New("no such database type: " + queryReq.DBType)
        return
    }
}

func (b *BasicAssetsApplication) PreExecute(req blockchainRequest.Entity, _ block.Entity) (result []byte, err error) {
    tx := basicAssetsLedger.Transaction{}
    err = json.Unmarshal(req.Data, &tx)
    if err != nil {
        return
    }

    if tx.TxType != req.RequestAction {
        err = errors.New("action not same as tx type")
        return
    }
    err = b.Ledger.VerifyTransaction(tx)

    return
}

func (b *BasicAssetsApplication) Execute(
        req blockchainRequest.Entity,
        blk block.Entity,
        actIndex uint32,
    ) (result applicationResult.Entity, err error) {

    tx := basicAssetsLedger.Transaction{}
    err = json.Unmarshal(req.Data, &tx)
    if err != nil {
        return
    }

    execResult, err := b.Ledger.ExecuteTransaction(tx)
    if err != nil {
        return
    }

    reqHash := req.Seal.Hash
    err =  b.Ledger.SaveTransactionWithBlockInfo(tx, reqHash, blk.Header.Height, actIndex)
    if err != nil {
        return
    }

    b.Ledger.RemoveTransactionFromPool(tx.HashString())

    if b.SQLStorage != nil {
        txWithBlk := basicAssetsLedger.TransactionWithBlockInfo {
            Transaction: tx,
        }

        txWithBlk.BlockInfo.RequestHash = reqHash
        txWithBlk.BlockInfo.BlockHeight = blk.Header.Height
        txWithBlk.BlockInfo.ActionIndex = actIndex

        txTypes :=  basicAssetsLedger.TransactionTypes

        switch tx.TxType {
        case txTypes.IssueAssets.String():
            b.storeAssets(txWithBlk, execResult)

        case txTypes.Transfer.String():
            b.storeTransfer(txWithBlk, execResult)

        case txTypes.StartSelling.String():
            fallthrough
        case txTypes.StopSelling.String():
            fallthrough
        case txTypes.BuyAssets.String():
            b.storeSelling(txWithBlk, execResult)

        case txTypes.IncreaseSupply.String():
            //todo: b.saveAssetsUpdate(txWithBlk)

        default:
            break
        }
    }
    return
}

func (b *BasicAssetsApplication) Cancel(req blockchainRequest.Entity) (err error) {
    return
}

func (b *BasicAssetsApplication) RequestsForBlock(_ block.Entity) (reqList []blockchainRequest.Entity, cnt uint32) {
    return b.Ledger.GetTransactionsFromPool()
}

func (b *BasicAssetsApplication) Information() (info service.BasicInformation) {
    info.Name = b.Name()
    info.Description = "this is a basic assets application based on a UTXO mode ledger"

    info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
    info.Api.Address = ""
    info.Api.ApiList = []service.ApiInterface {}
    return
}

func (b *BasicAssetsApplication) SetChainInterface(_ chainStructure.IChainInterface) {}
func (b *BasicAssetsApplication) UnpackingActionsAsRequests(_ blockchainRequest.Entity) ([]blockchainRequest.Entity, error) {return nil, nil}

func (b *BasicAssetsApplication) GetActionAsRequest(req blockchainRequest.Entity) blockchainRequest.Entity {
    tx := basicAssetsLedger.Transaction{}
    _ = json.Unmarshal(req.Data, &tx)

    newReq := blockchainRequest.Entity{}
    newReq.Seal = tx.Seal
    newReq.RequestApplication = b.Name()
    newReq.RequestAction = tx.TxType

    newReq.Data, _ = json.Marshal(tx.TransactionData)

    return newReq
}

func Load()  {
    enum.SimpleBuild(&QueryDBType)
    basicAssetsLedger.Load()
    basicAssetsSQLStorage.Load()
}

func NewApplicationInterface(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver) (app chainStructure.IBlockchainExternalApplication) {
    bs := BasicAssetsApplication{}
    bs.Ledger = basicAssetsLedger.NewLedger(kvDriver)
    if sqlDriver != nil {
        bs.SQLStorage = basicAssetsSQLStorage.NewStorage(sqlDriver)
    }

    app = &bs
    return
}
