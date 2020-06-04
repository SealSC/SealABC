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

package smartAssetsSQLTables

import (
	"SealABC/common"
	"SealABC/dataStructure/enum"
	"SealABC/metadata/block"
	"SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

type AddressListTable struct {
	ID           enum.Element `col:"c_id" ignoreInsert:"true"`
	Address      enum.Element `col:"c_address"`
	Balance      enum.Element `col:"c_balance"`
	UpdateHeight enum.Element `col:"c_update_height"`
	UpdateTime   enum.Element `col:"c_update_time"`

	simpleSQLDatabase.BasicTable
}

var AddressList AddressListTable

func (a AddressListTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(AddressListRows{})
}

func (a AddressListTable) Name() (name string) {
	return "t_smart_assets_address_list"
}

func (a *AddressListTable) load() {
	enum.SimpleBuild(a)
	a.Instance = *a
}

type AddressListRow struct {
	ID           string
	Address      string
	Balance      string
	UpdateHeight string
	UpdateTime   string
}

type AddressListRows struct {
	simpleSQLDatabase.BasicRows
}

func (a *AddressListRows) Insert(addr []byte, balance *big.Int, blk block.Entity) {
	timestamp := time.Unix(int64(blk.Header.Timestamp), 0)
	newAddressRow := AddressListRow{
		Address:      hex.EncodeToString(addr),
		Balance:      balance.String(),
		UpdateHeight: fmt.Sprintf("%d", blk.Header.Height),
		UpdateTime:   timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	a.Rows = append(a.Rows, newAddressRow)
}

func (a *AddressListRows) InsertSystemIssueBalance(balance *big.Int, address string)  {
	newRow := AddressListRow{
		Address:      address,
		Balance:      balance.String(),
		UpdateHeight: "0",
		UpdateTime:   time.Now().Format(common.BASIC_TIME_FORMAT),
	}
	a.Rows = append(a.Rows, newRow)
}

func (a *AddressListRows) Table() simpleSQLDatabase.ITable {
	return &AddressList
}
