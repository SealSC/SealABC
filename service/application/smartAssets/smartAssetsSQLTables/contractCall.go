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
	"encoding/json"
	"fmt"
	"time"
)

type ContractCallTable struct {
	ID              enum.Element `col:"c_id" ignoreInsert:"true"`
	Height          enum.Element `col:"c_height"`
	TxHash          enum.Element `col:"c_tx_hash"`
	SequenceNumber  enum.Element `col:"c_sequence_number"`
	Caller          enum.Element `col:"c_caller"`
	ContractAddress enum.Element `col:"c_contract_address"`
	Data            enum.Element `col:"c_data"`
	Result          enum.Element `col:"c_result"`
	Value           enum.Element `col:"c_value"`
	Time            enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var ContractCall ContractCallTable

func (t ContractCallTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(ContractCallRows{})
}

func (t ContractCallTable) Name() (name string) {
	return "t_smart_assets_contract_call"
}

func (t *ContractCallTable) load() {
	enum.SimpleBuild(t)
	t.Instance = *t
}

type ContractCallRow struct {
	ID              string
	Height          string
	TxHash          string
	SequenceNumber  string
	Caller          string
	ContractAddress string
	Data            string
	Result          string
	Value           string
	Time            string
}

type ContractCallRows struct {
	simpleSQLDatabase.BasicRows
}

func (t *ContractCallRows) Insert(tx smartAssetsLedger.Transaction, blk block.Entity) {
	timestamp := time.Unix(int64(blk.Header.Timestamp), 0)
	result, _ := json.Marshal(tx.TransactionResult)
	newAddressRow := ContractCallRow{
		Height:          fmt.Sprintf("%d", blk.Header.Height),
		TxHash:          hex.EncodeToString(tx.DataSeal.Hash),
		SequenceNumber:  fmt.Sprintf("%d", tx.SequenceNumber),
		Caller:          hex.EncodeToString(tx.From),
		ContractAddress: hex.EncodeToString(tx.To),
		Data:            hex.EncodeToString(tx.Data),
		Result:          string(result),
		Value:           tx.Value,
		Time:            timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	t.Rows = append(t.Rows, newAddressRow)
}

func (t *ContractCallRows) Table() simpleSQLDatabase.ITable {
	return &ContractCall
}
