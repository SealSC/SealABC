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

package basicAssetsTables

import (
    "SealABC/common"
    "SealABC/dataStructure/enum"
    "SealABC/service/application/basicAssets/basicAssetsLedger"
    "SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "encoding/hex"
    "fmt"
    "time"
)

type AssetsListTable struct {
    ID              enum.Element `col:"c_id" ignoreInsert:"true"`
    Height          enum.Element `col:"c_height"`
    ReqHash         enum.Element `col:"c_req_hash"`
    TxHash          enum.Element `col:"c_tx_hash"`
    AssetsName      enum.Element `col:"c_name"`
    AssetsSymbol    enum.Element `col:"c_symbol"`
    Supply          enum.Element `col:"c_supply"`
    Increasable     enum.Element `col:"c_increasable"`
    MetaHash        enum.Element `col:"c_meta_hash"`
    MetaSignature   enum.Element `col:"c_meta_signature"`
    IssuedHash      enum.Element `col:"c_issued_hash"`
    IssuedSignature enum.Element `col:"c_issued_signature"`
    IssueTo         enum.Element `col:"c_issue_to"`
    Time            enum.Element `col:"c_time"`

    simpleSQLDatabase.BasicTable
}

var AssetsList AssetsListTable

func (a AssetsListTable) Name() (name string) {
    return "t_basic_assets_list"
}

func (a AssetsListTable) NewRows() interface{} {
    return simpleSQLDatabase.NewRowsInstance(AssetsListRows{})
}


func (a *AssetsListTable) load() {
    enum.SimpleBuild(a)
    a.Instance = *a
}

type AssetsListRow struct {
    ID              string
    Height          string
    ReqHash         string
    TxHash          string
    AssetsName      string
    AssetsSymbol    string
    Supply          string
    Increasable     string
    MetaHash        string
    MetaSignature   string
    IssuedHash      string
    IssuedSignature string
    IssueTo         string
    Time            string
}

func (b *AssetsListRow) FromTransaction(tx basicAssetsLedger.TransactionWithBlockInfo)  {
    b.Height = fmt.Sprintf("%d", tx.BlockInfo.BlockHeight)
    b.ReqHash = hex.EncodeToString(tx.BlockInfo.RequestHash)
    b.TxHash = hex.EncodeToString(tx.Seal.Hash)
    b.AssetsName = tx.Assets.Name
    b.AssetsSymbol = tx.Assets.Symbol
    b.Supply = fmt.Sprintf("%d", tx.Assets.Supply)
    if tx.Assets.Increasable {
        b.Increasable = "1"
    } else {
        b.Increasable = "0"
    }

    b.MetaHash = hex.EncodeToString(tx.Assets.MetaSeal.Hash)
    b.MetaSignature = hex.EncodeToString(tx.Assets.MetaSeal.Signature)

    b.IssuedHash = hex.EncodeToString(tx.Assets.IssuedSeal.Hash)
    b.IssuedSignature = hex.EncodeToString(tx.Assets.IssuedSeal.Signature)

    b.IssueTo = hex.EncodeToString(tx.Assets.MetaSeal.SignerPublicKey)

    timestamp := time.Unix(int64(tx.CreateTime), 0)
    b.Time = timestamp.Format(common.BASIC_TIME_FORMAT)

    return
}

type AssetsListRows struct {
    simpleSQLDatabase.BasicRows
}

func (b *AssetsListRows) InsertAssets(issueTX basicAssetsLedger.TransactionWithBlockInfo)  {
    newRow := AssetsListRow{}
    newRow.FromTransaction(issueTX)
    b.Rows = append(b.Rows, newRow)
}

func (b *AssetsListRows) Table() simpleSQLDatabase.ITable {
    return &AssetsList
}
