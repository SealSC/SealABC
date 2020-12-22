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
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsLedger"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"encoding/hex"
	"fmt"
	"time"
)

type SellingListTable struct {
	ID               enum.Element `col:"c_id" ignoreInsert:"true"`
	Height           enum.Element `col:"c_height"`
	StartTransaction enum.Element `col:"c_start_transaction"`
	StopTransaction  enum.Element `col:"c_stop_transaction"`
	Seller           enum.Element `col:"c_seller"`
	Buyer            enum.Element `col:"c_buyer"`
	SellingAssets    enum.Element `col:"c_selling_assets"`
	PaymentAssets    enum.Element `col:"c_payment_assets"`
	Price            enum.Element `col:"c_price"`
	Status           enum.Element `col:"c_status"`
	StartTime        enum.Element `col:"c_start_time"`
	StopTime         enum.Element `col:"c_stop_time"`

	simpleSQLDatabase.BasicTable
}

var SellingList SellingListTable

func (s SellingListTable) Name() (name string) {
	return "t_basic_assets_selling_list"
}

func (s SellingListTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(SellingListRows{})
}


func (s *SellingListTable) load() {
	enum.SimpleBuild(s)
	s.Instance = *s
}

type SellingListRow struct {
	ID               string
	Height           string
	StartTransaction string
	StopTransaction  string
	Seller           string
	Buyer            string
	SellingAssets    string
	PaymentAssets    string
	Price            string
	Status           string
	StartTime        string
	StopTime         string
}

func (s *SellingListRow) FromTransaction(tx basicAssetsLedger.TransactionWithBlockInfo, sellingData basicAssetsLedger.SellingData)  {
	s.Height = fmt.Sprintf("%d", tx.BlockInfo.BlockHeight)

	s.Seller = hex.EncodeToString(sellingData.Seller)
	s.SellingAssets = hex.EncodeToString(sellingData.SellingAssets)
	s.PaymentAssets = hex.EncodeToString(sellingData.PaymentAssets)
	s.Price = fmt.Sprintf("%d", sellingData.Price)

	timestamp := time.Unix(tx.CreateTime, 0)
	timeString := timestamp.Format(common.BASIC_TIME_FORMAT)

	if tx.TxType == basicAssetsLedger.TransactionTypes.StartSelling.String() {
		s.StartTransaction = hex.EncodeToString(tx.Seal.Hash)
		s.Status = "0"
		s.StartTime = timeString
		s.StopTime = timeString
	} else {
		s.StopTransaction = hex.EncodeToString(tx.Seal.Hash)
		s.StopTime = timeString
		if tx.TxType == basicAssetsLedger.TransactionTypes.StopSelling.String() {
			s.Status = "2"
		} else {
			s.Status = "1"
			s.Buyer = hex.EncodeToString(tx.Seal.SignerPublicKey)
		}
	}
	return
}

type SellingListRows struct {
	simpleSQLDatabase.BasicRows
}

func (s *SellingListRows) InsertRow(tx basicAssetsLedger.TransactionWithBlockInfo, data basicAssetsLedger.SellingData) {
	newRow := SellingListRow{}
	newRow.FromTransaction(tx, data)
	s.Rows = append(s.Rows, newRow)
}

func (s *SellingListRows) GetUpdateInfo() ([]string, string) {
	return []string{
		"Height",
		"StopTransaction",
		"Buyer",
		"Status",
		"StopTime",
	},
	"where c_start_transaction=?"
}

func (s *SellingListRows) Table() simpleSQLDatabase.ITable {
	return &SellingList
}
