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

package memoSQLStorage

import (
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/service/application/memo/memoSpace"
	"github.com/SealSC/SealABC/service/application/memo/memoTables"
)

func (s *Storage) StoreMemo(height uint64, tm int64, req blockchainRequest.Entity, memo memoSpace.Memo) (err error) {
	memoListRows := memoTables.MemoList.NewRows().(memoTables.MemoListRows)
	memoListRows.InsertMemo(height, tm, req, memo)

	_, err = s.Driver.Insert(&memoListRows, true)
	if err != nil {
		log.Log.Error("insert memo to sql database failed: ", err.Error())
	}

	address := memo.Seal.HexPublicKey()
	addressListRows := memoTables.AddressList.NewRows().(memoTables.AddressListRows)
	addressListRows.InsertAddress(height, tm, address)
	_, err = s.Driver.Insert(&addressListRows, true)
	if err != nil {
		log.Log.Error("insert memo address to sql database failed: ", err.Error())
	}
	return
}
