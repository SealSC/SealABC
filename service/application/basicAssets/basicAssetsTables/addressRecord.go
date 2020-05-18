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

var AddressRoles struct{
    Issuer  enum.Element
    Payer   enum.Element
    Payee   enum.Element
}

type AddressRecordTable struct {
    ID      enum.Element `col:"c_id" ignoreInsert:"true"`
    Address enum.Element `col:"c_address"`
    Height  enum.Element `col:"c_height"`
    TxHash  enum.Element `col:"c_tx_hash"`
    Role    enum.Element `col:"c_tx_role"`
    Time    enum.Element `col:"c_time"`

    simpleSQLDatabase.BasicTable
}

var AddressRecord AddressRecordTable

func (a AddressRecordTable) Name() (name string) {
    return "t_basic_assets_address_record"
}

func (a AddressRecordTable) NewRows() interface{} {
    return simpleSQLDatabase.NewRowsInstance(AddressRecordRows{})
}

func (a *AddressRecordTable) load() {
    enum.SimpleBuild(a)
    enum.SimpleBuild(&AddressRoles)

    a.Instance = *a
}

type AddressRecordRow struct {
    ID      string
    Address string
    Height  string
    TxHash  string
    Role    string
    Time    string
}

type AddressRecordRows struct {
    simpleSQLDatabase.BasicRows
}

func (b *AddressRecordRows) InsertAddress(tx basicAssetsLedger.TransactionWithBlockInfo, address string, role enum.Element)  {
    timestamp := time.Unix(tx.CreateTime, 0)
    newAddressRow := AddressRecordRow{
        Address: address,
        Height:  fmt.Sprintf("%d", tx.BlockInfo.BlockHeight),
        TxHash:  hex.EncodeToString(tx.Seal.Hash),
        Role:    role.String(),
        Time:    timestamp.Format(common.BASIC_TIME_FORMAT),
    }

    b.Rows = append(b.Rows, newAddressRow)
}

func (b *AddressRecordRows) InsertAddressesInTransfer(tx basicAssetsLedger.TransactionWithBlockInfo, unspentList []basicAssetsLedger.Unspent)  {
    for _, in := range unspentList {
        b.InsertAddress(tx, hex.EncodeToString(in.Owner), AddressRoles.Payer)
    }

    for _, out := range tx.Output {
        b.InsertAddress(tx, hex.EncodeToString(out.To), AddressRoles.Payee)
    }
}

func (b *AddressRecordRows) Table() simpleSQLDatabase.ITable {
    return &AddressRecord
}
