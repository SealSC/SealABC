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
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsLedger"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"time"
)

type AddressListTable struct {
	ID      enum.Element `col:"c_id" ignoreInsert:"true"`
	Height  enum.Element `col:"c_height"`
	Address enum.Element `col:"c_address"`
	Time    enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var AddressList AddressListTable

func (a AddressListTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(AddressListRows{})
}

func (a AddressListTable) Name() (name string) {
	return "t_basic_assets_address_list"
}

func (a *AddressListTable) load() {
	enum.SimpleBuild(a)
	a.Instance = *a
}

type AddressListRow struct {
	ID      string
	Height  string
	Address string
	Time    string
}

type AddressListRows struct {
	simpleSQLDatabase.BasicRows
}

func (b *AddressListRows) InsertAddress(tx basicAssetsLedger.TransactionWithBlockInfo, address string) {
	timestamp := time.Unix(tx.CreateTime, 0)
	newAddressRow := AddressListRow{
		Address: address,
		Height:  fmt.Sprintf("%d", tx.BlockInfo.BlockHeight),
		Time:    timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	b.Rows = append(b.Rows, newAddressRow)
}

func (b *AddressListRows) Table() simpleSQLDatabase.ITable {
	return &AddressList
}
