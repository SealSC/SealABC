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
	"SealABC/service/application/smartAssets/smartAssetsLedger"
	"SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"encoding/hex"
	"fmt"
	"time"
)

type ContractTable struct {
	ID              enum.Element `col:"c_id" ignoreInsert:"true"`
	Height          enum.Element `col:"c_height"`
	TxHash          enum.Element `col:"c_tx_hash"`
	SequenceNumber  enum.Element `col:"c_sequence_number"`
	Creator         enum.Element `col:"c_creator"`
	CreationData    enum.Element `col:"c_creation_data"`
	ContractAddress enum.Element `col:"c_contract_address"`
	ContractData    enum.Element `col:"c_contract_data"`
	Time            enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var Contract ContractTable

func (t ContractTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(ContractRows{})
}

func (t ContractTable) Name() (name string) {
	return "t_smart_assets_contract"
}

func (t *ContractTable) load() {
	enum.SimpleBuild(t)
	t.Instance = *t
}

type ContractRow struct {
	ID              string
	Height          string
	TxHash          string
	SequenceNumber  string
	Creator         string
	CreationData    string
	ContractAddress string
	ContractData    string
	Time            string
}

type ContractRows struct {
	simpleSQLDatabase.BasicRows
}

func (t *ContractRows) Insert(tx smartAssetsLedger.Transaction, blk block.Entity) {
	timestamp := time.Unix(int64(blk.Header.Timestamp), 0)
	newAddressRow := ContractRow{
		Height:          fmt.Sprintf("%d", blk.Header.Height),
		TxHash:          hex.EncodeToString(tx.DataSeal.Hash),
		SequenceNumber:  fmt.Sprintf("%d", tx.SequenceNumber),
		Creator:         hex.EncodeToString(tx.From),
		CreationData:    hex.EncodeToString(tx.Data),
		ContractAddress: hex.EncodeToString(tx.TransactionResult.NewAddress),
		ContractData:    hex.EncodeToString(tx.TransactionResult.ReturnData),
		Time:            timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	t.Rows = append(t.Rows, newAddressRow)
}

func (t *ContractRows) Table() simpleSQLDatabase.ITable {
	return &Contract
}
