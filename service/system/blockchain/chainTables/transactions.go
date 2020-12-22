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
    "github.com/SealSC/SealABC/common"
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/metadata/applicationResult"
    "github.com/SealSC/SealABC/metadata/block"
    "github.com/SealSC/SealABC/metadata/blockchainRequest"
    "github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
)

type RequestsTable struct {
   ID           enum.Element `col:"c_id" ignoreInsert:"true"`
   Height       enum.Element `col:"c_height"`
   Sender       enum.Element `col:"c_sender"`
   Hash         enum.Element `col:"c_hash"`
   Sig          enum.Element `col:"c_signature"`
   Application  enum.Element `col:"c_application"`
   Action       enum.Element `col:"c_action"`
   Payload      enum.Element `col:"c_payload"`
   Packed       enum.Element `col:"c_packed"`
   Result       enum.Element `col:"c_result"`
   Time         enum.Element `col:"c_time"`

   simpleSQLDatabase.BasicTable
}

var Requests RequestsTable

func (t RequestsTable) Name() (name string) {
    return "t_block_requests"
}

func (t RequestsTable) NewRows() interface{} {
    return simpleSQLDatabase.NewRowsInstance(RequestRows{})
}

func (t *RequestsTable) load() {
    enum.SimpleBuild(t)
    t.Instance = *t
}

type RequestRow struct {
    ID          string
    Height      string
    Sender      string
    Hash        string
    Sig         string
    Application string
    Action      string
    Payload     string
    Packed      string
    Result      string
    Time        string
}

func (t *RequestRow) FromRequest(height uint64, tm uint64, req blockchainRequest.Entity, result applicationResult.Entity) {
   t.Height = fmt.Sprintf("%d", height)
   t.Sender = hex.EncodeToString(req.Seal.SignerPublicKey)
   t.Hash = hex.EncodeToString(req.Seal.Hash)
   t.Sig = hex.EncodeToString(req.Seal.Signature)
   t.Application = req.RequestApplication
   t.Action = req.RequestAction

   reqJson, _ := json.Marshal(req.EntityData)
   t.Payload = string(reqJson)

   if req.Packed {
       t.Packed = "1"
   } else {
       t.Packed = "0"
   }

   resultJson, _ := json.Marshal(result)
   t.Result = string(resultJson)

   timestamp := time.Unix(int64(tm), 0)
   t.Time = timestamp.Format(common.BASIC_TIME_FORMAT)
}

type RequestRows struct {
    simpleSQLDatabase.BasicRows
}

func (t *RequestRows) InsertTransaction(blk block.Entity, req blockchainRequest.Entity, result applicationResult.Entity)  {
    newRow := RequestRow{}
    newRow.FromRequest(blk.Header.Height, blk.Header.Timestamp, req, result)
    t.Rows = append(t.Rows, newRow)
}

func (t *RequestRows) Table() simpleSQLDatabase.ITable {
    return &Requests
}
