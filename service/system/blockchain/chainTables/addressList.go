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

package chainTables

import (
	"encoding/hex"
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"time"
)

type AddressListTable struct {
	Address     enum.Element `col:"c_address"`
	Height      enum.Element `col:"c_last_height"`
	Transaction enum.Element `col:"c_last_transaction"`
	Application enum.Element `col:"c_last_application"`
	Action      enum.Element `col:"c_last_action"`
	Time        enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var AddressList AddressListTable

func (t AddressListTable) Name() (name string) {
	return "t_address_list"
}

func (t AddressListTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(AddressListRows{})
}

func (t *AddressListTable) load() {
	enum.SimpleBuild(t)
	t.Instance = *t
}

type AddressListRow struct {
	Address     string
	Height      string
	Transaction string
	Application string
	Action      string
	Time        string
}

func (t *AddressListRow) FromRequest(height uint64, tm uint64, req blockchainRequest.Entity) {
	t.Height = fmt.Sprintf("%d", height)
	t.Address = hex.EncodeToString(req.Seal.SignerPublicKey)
	t.Transaction = hex.EncodeToString(req.Seal.Hash)
	t.Application = req.RequestApplication
	t.Action = req.RequestAction

	timestamp := time.Unix(int64(tm), 0)
	t.Time = timestamp.Format(common.BASIC_TIME_FORMAT)
}

type AddressListRows struct {
	simpleSQLDatabase.BasicRows
}

func (t *AddressListRows) UpdateAddress(blk block.Entity, req blockchainRequest.Entity) {
	newRow := AddressListRow{}
	newRow.FromRequest(blk.Header.Height, blk.Header.Timestamp, req)
	t.Rows = append(t.Rows, newRow)
}

func (t *AddressListRows) Table() simpleSQLDatabase.ITable {
	return &AddressList
}
