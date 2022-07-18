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
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/service/application/memo/memoSpace"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"time"
)

type MemoListTable struct {
	ID       enum.Element `col:"c_id" ignoreInsert:"true"`
	Height   enum.Element `col:"c_height"`
	ReqHash  enum.Element `col:"c_req_hash"`
	Recorder enum.Element `col:"c_recorder"`
	Type     enum.Element `col:"c_type"`
	Data     enum.Element `col:"c_data"`
	Hash     enum.Element `col:"c_hash"`
	Sig      enum.Element `col:"c_signature"`
	Size     enum.Element `col:"c_size"`
	Time     enum.Element `col:"c_time"`

	simpleSQLDatabase.BasicTable
}

var MemoList MemoListTable

func (m MemoListTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(MemoListRows{})
}

func (m MemoListTable) Name() (name string) {
	return "t_memo_list"
}

func (m *MemoListTable) load() {
	enum.SimpleBuild(m)
	m.Instance = *m
}

type MemoListRow struct {
	ID       string
	Height   string
	ReqHash  string
	Recorder string
	Type     string
	Data     string
	Hash     string
	Sig      string
	Size     string
	Time     string
}

type MemoListRows struct {
	simpleSQLDatabase.BasicRows
}

func (m *MemoListRows) InsertMemo(height uint64, tm int64, tx blockchainRequest.Entity, memo memoSpace.Memo) {
	timestamp := time.Unix(tm, 0)

	memoDataSize := len(memo.Data)

	newAddressRow := MemoListRow{
		Height:   fmt.Sprintf("%d", height),
		ReqHash:  tx.Seal.HexHash(),
		Recorder: memo.Seal.HexPublicKey(),
		Type:     memo.Type,
		Data:     memo.Data,
		Hash:     memo.Seal.HexHash(),
		Sig:      memo.Seal.HexSignature(),
		Size:     fmt.Sprintf("%d", memoDataSize),
		Time:     timestamp.Format(common.BASIC_TIME_FORMAT),
	}
	m.Rows = append(m.Rows, newAddressRow)
}

func (m *MemoListRows) Table() simpleSQLDatabase.ITable {
	return &MemoList
}
