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
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsLedger"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"encoding/hex"
	"fmt"
	"time"
)

type TransferTable struct {
	ID             enum.Element `col:"c_id" ignoreInsert:"true"`
	Height         enum.Element `col:"c_height"`
	TxHash         enum.Element `col:"c_tx_hash"`
	SequenceNumber enum.Element `col:"c_sequence_number"`
	From           enum.Element `col:"c_from"`
	To             enum.Element `col:"c_to"`
	Value          enum.Element `col:"c_value"`
	Memo          enum.Element `col:"c_memo"`
	Time           enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var Transfer TransferTable

func (t TransferTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(TransferRows{})
}

func (t TransferTable) Name() (name string) {
	return "t_smart_assets_transfer"
}

func (t *TransferTable) load() {
	enum.SimpleBuild(t)
	t.Instance = *t
}

type TransferRow struct {
	ID             string
	Height         string
	TxHash         string
	SequenceNumber string
	From           string
	To             string
	Value          string
	Memo           string
	Time           string
}

type TransferRows struct {
	simpleSQLDatabase.BasicRows
}

func (t *TransferRows) Insert(tx smartAssetsLedger.Transaction, blk block.Entity) {
	timestamp := time.Unix(int64(blk.Header.Timestamp), 0)
	newAddressRow := TransferRow{
		Height:         fmt.Sprintf("%d", blk.Header.Height),
		TxHash:         hex.EncodeToString(tx.DataSeal.Hash),
		SequenceNumber: fmt.Sprintf("%d", tx.SequenceNumber),
		From:           hex.EncodeToString(tx.From),
		To:             hex.EncodeToString(tx.To),
		Value:          tx.Value,
		Memo:           tx.Memo,
		Time:           timestamp.Format(common.BASIC_TIME_FORMAT),
	}

	t.Rows = append(t.Rows, newAddressRow)
}

func (t *TransferRows) Table() simpleSQLDatabase.ITable {
	return &Transfer
}
