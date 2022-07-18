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

package memoTables

import (
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/enum"
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
	return "t_memo_address_list"
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

func (b *AddressListRows) InsertAddress(height uint64, tm int64, address string) {
	timestamp := time.Unix(tm, 0)
	newAddressRow := AddressListRow{
		Height:  fmt.Sprintf("%d", height),
		Address: address,
		Time:    timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	b.Rows = append(b.Rows, newAddressRow)
}

func (b *AddressListRows) Table() simpleSQLDatabase.ITable {
	return &AddressList
}
