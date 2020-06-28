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

type TransactionTable struct {
	ID             enum.Element `col:"c_id" ignoreInsert:"true"`
	Height         enum.Element `col:"c_height"`
	TxHash         enum.Element `col:"c_hash"`
	Type           enum.Element `col:"c_type"`
	From           enum.Element `col:"c_from"`
	To             enum.Element `col:"c_to"`
	Value          enum.Element `col:"c_value"`
	Data           enum.Element `col:"c_data"`
	Memo           enum.Element `col:"c_memo"`
	SerialNumber   enum.Element `col:"c_serial_number"`
	SequenceNumber enum.Element `col:"c_sequence_number"`
	TxDataSeal     enum.Element `col:"c_tx_data_seal"`
	Result         enum.Element `col:"c_result"`
	Time           enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var Transaction TransactionTable

func (t TransactionTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(TransactionRows{})
}

func (t TransactionTable) Name() (name string) {
	return "t_smart_assets_transaction"
}

func (t *TransactionTable) load() {
	enum.SimpleBuild(t)
	t.Instance = *t
}

type TransactionRow struct {
	ID             string
	Height         string
	TxHash         string
	Type           string
	From           string
	To             string
	Value          string
	Data           string
	Memo           string
	SerialNumber   string
	SequenceNumber string
	TxDataSeal     string
	Result         string
	Time           string
}

type TransactionRows struct {
	simpleSQLDatabase.BasicRows
}

func (t *TransactionRows) Insert(tx smartAssetsLedger.Transaction, blk block.Entity) {
	timestamp := time.Unix(int64(blk.Header.Timestamp), 0)
	sealData, _ := json.Marshal(tx.DataSeal)
	result, _ := json.Marshal(tx.TransactionResult)
	newAddressRow := TransactionRow{
		Height:         fmt.Sprintf("%d", blk.Header.Height),
		TxHash:         hex.EncodeToString(tx.DataSeal.Hash),
		Type:           tx.Type,
		From:           hex.EncodeToString(tx.From),
		To:             hex.EncodeToString(tx.To),
		Value:          tx.Value,
		Data:           hex.EncodeToString(tx.Data),
		Memo:           tx.Memo,
		SerialNumber:   tx.SerialNumber,
		SequenceNumber: fmt.Sprintf("%d", tx.SequenceNumber),
		TxDataSeal:     string(sealData),
		Result:         string(result),
		Time:           timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	t.Rows = append(t.Rows, newAddressRow)
}

func (t *TransactionRows) Table() simpleSQLDatabase.ITable {
	return &Transaction
}

