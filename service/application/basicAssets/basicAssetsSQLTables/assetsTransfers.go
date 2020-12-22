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

package basicAssetsSQLTables

import (
    "github.com/SealSC/SealABC/common"
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsLedger"
    "github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
)

type TransfersTable struct {
    ID      enum.Element `col:"c_id" ignoreInsert:"true"`
    Height  enum.Element `col:"c_height"`
    ReqHash enum.Element `col:"c_req_hash"`
    TxHash  enum.Element `col:"c_tx_hash"`
    From    enum.Element `col:"c_from"`
    To      enum.Element `col:"c_to"`
    Assets  enum.Element `col:"c_assets_hash"`
    Amount  enum.Element `col:"c_amount"`
    Type    enum.Element `col:"c_transaction_type"`
    Memo    enum.Element `col:"c_transaction_memo"`
    Time    enum.Element `col:"c_time"`

    simpleSQLDatabase.BasicTable
}

var Transfers TransfersTable

func (b TransfersTable) Name() (name string) {
    return "t_basic_assets_transfers"
}

func (b TransfersTable) NewRows() interface{} {
    return simpleSQLDatabase.NewRowsInstance(TransfersRows{})
}

func (b *TransfersTable) load() {
    enum.SimpleBuild(b)
    b.Instance = *b
}


type TransfersRow struct {
    ID      string
    Height  string
    ReqHash string
    TxHash  string
    From    string
    To      string
    Assets  string
    Amount  string
    Type    string
    Memo    string
    Time    string
}

type TransfersRows struct {
    simpleSQLDatabase.BasicRows
}

func (b *TransfersRows) InsertTransferInsideIssueTransaction(tx basicAssetsLedger.TransactionWithBlockInfo)  {
    fromJson, _ := json.Marshal(map[string] uint64 {"blockchain": tx.Assets.Supply})
    toJson, _ := json.Marshal(map[string] uint64 {
        hex.EncodeToString(tx.Assets.MetaSeal.SignerPublicKey): tx.Assets.Supply,
    })

    timestamp := time.Unix(tx.CreateTime, 0)
    newTransfersRow := TransfersRow{
        Height: fmt.Sprintf("%d", tx.BlockInfo.BlockHeight),
        ReqHash: hex.EncodeToString(tx.BlockInfo.RequestHash),
        TxHash: hex.EncodeToString(tx.Seal.Hash),
        From: string(fromJson),
        To: string(toJson),
        Assets: hex.EncodeToString(tx.Assets.MetaSeal.Hash),
        Amount: fmt.Sprintf("%d", tx.Assets.Supply),
        Type: tx.TxType,
        Memo: tx.Memo,
        Time: timestamp.Format(common.BASIC_TIME_FORMAT),
    }

    b.Rows = append(b.Rows, newTransfersRow)
}

func (b *TransfersRows) InsertTransfer(tx basicAssetsLedger.TransactionWithBlockInfo, unspentList []basicAssetsLedger.Unspent)  {
    from := map[string] uint64 {}
    for _, in := range unspentList {
        ownerKey := hex.EncodeToString(in.Owner)
        from[ownerKey] += in.Value
    }
    fromJson, _ := json.Marshal(from)

    to := map[string] uint64 {}
    var amount uint64 = 0
    for _, out := range tx.Output {
        outKey := hex.EncodeToString(out.To)
        to[outKey] += out.Value
        amount += out.Value
    }
    toJson, _ := json.Marshal(to)

    timestamp := time.Unix(tx.CreateTime, 0)
    newTransfersRow := TransfersRow{
        Height: fmt.Sprintf("%d", tx.BlockInfo.BlockHeight),
        ReqHash: hex.EncodeToString(tx.BlockInfo.RequestHash),
        TxHash: hex.EncodeToString(tx.Seal.Hash),
        From: string(fromJson),
        To: string(toJson),
        Assets: hex.EncodeToString(tx.Assets.MetaSeal.Hash),
        Amount: fmt.Sprintf("%d", amount),
        Type: tx.TxType,
        Memo: tx.Memo,
        Time: timestamp.Format(common.BASIC_TIME_FORMAT),
    }

    b.Rows = append(b.Rows, newTransfersRow)
}

func (b *TransfersRows) InsertTransferByDetail(
    tx basicAssetsLedger.TransactionWithBlockInfo,
    assets []byte,
    input  []basicAssetsLedger.Unspent,
    output []basicAssetsLedger.UTXOOutput) {

    var amount uint64 = 0
    from := map[string] uint64 {}
    for _, in := range input {
        ownerKey := hex.EncodeToString(in.Owner)
        from[ownerKey] += in.Value
    }
    fromJson, _ := json.Marshal(from)

    to := map[string] uint64 {}
    for _, o := range output {
        outKey := hex.EncodeToString(o.To)
        to[outKey] += o.Value
        amount += o.Value
    }

    toJson, _ := json.Marshal(to)

    timestamp := time.Unix(tx.CreateTime, 0)

    newRow := TransfersRow {
        Height: fmt.Sprintf("%d", tx.BlockInfo.BlockHeight),
        ReqHash: hex.EncodeToString(tx.BlockInfo.RequestHash),
        TxHash: hex.EncodeToString(tx.Seal.Hash),
        From:    string(fromJson),
        To:      string(toJson),
        Assets:  hex.EncodeToString(assets),
        Amount:  fmt.Sprintf("%d", amount),
        Type:    tx.TxType,
        Memo:    tx.Memo,
        Time:    timestamp.Format(common.BASIC_TIME_FORMAT),
    }

    b.Rows = append(b.Rows, newRow)
}

func (b *TransfersRows) Table() simpleSQLDatabase.ITable {
    return &Transfers
}
