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
	"encoding/json"
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"time"
)

type BlockListTable struct {
	ID       enum.Element `col:"c_id" ignoreInsert:"true"`
	Height   enum.Element `col:"c_height"`
	Signer   enum.Element `col:"c_signer"`
	Hash     enum.Element `col:"c_hash"`
	Sig      enum.Element `col:"c_signature"`
	PrevHash enum.Element `col:"c_prev_hash"`
	Time     enum.Element `col:"c_time"`
	TXRoot   enum.Element `col:"c_tx_root" def:"-"`
	TXCount  enum.Element `col:"c_tx_count" def:"0"`
	Payload  enum.Element `col:"c_payload" def:"{}"`

	simpleSQLDatabase.BasicTable
}

var BlockList BlockListTable

func (b BlockListTable) Name() (name string) {
	return "t_block_list"
}

func (b BlockListTable) NewRows() interface{} {
	return simpleSQLDatabase.NewRowsInstance(BlockListRows{})
}

func (b *BlockListTable) load() {
	enum.SimpleBuild(b)
	b.Instance = *b
}

type BlockListRow struct {
	ID       string
	Height   string
	Signer   string
	Hash     string
	Sig      string
	PrevHash string
	Time     string
	TXRoot   string
	TXCount  string
	Payload  string
}

func (b *BlockListRow) FromBlockEntity(blk block.Entity) {
	b.Height = fmt.Sprintf("%d", blk.Header.Height)
	b.Signer = hex.EncodeToString(blk.Seal.SignerPublicKey)
	b.Hash = hex.EncodeToString(blk.Seal.Hash)
	b.Sig = hex.EncodeToString(blk.Seal.Signature)
	b.PrevHash = hex.EncodeToString(blk.Header.PrevBlock)

	timestamp := time.Unix(int64(blk.Header.Timestamp), 0)
	b.Time = timestamp.Format(common.BASIC_TIME_FORMAT)
	b.TXRoot = hex.EncodeToString(blk.Header.TransactionsRoot)
	b.TXCount = fmt.Sprintf("%d", blk.Body.RequestsCount)

	payloadJson, _ := json.Marshal(blk.EntityData)
	b.Payload = string(payloadJson)
}

type BlockListRows struct {
	simpleSQLDatabase.BasicRows
}

func (b *BlockListRows) InsertBlock(blk block.Entity) {
	newRow := BlockListRow{}
	newRow.FromBlockEntity(blk)
	b.Rows = append(b.Rows, newRow)
}

func (b *BlockListRows) Table() simpleSQLDatabase.ITable {
	return &BlockList
}
