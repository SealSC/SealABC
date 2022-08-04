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

package memoInterface

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/service/application/memo/memoSpace"
)

func (m *MemoApplication) VerifyReq(req blockchainRequest.Entity) (passed bool, memo memoSpace.Memo, err error) {
	err = json.Unmarshal(req.Data, &memo)
	if err != nil {
		log.Log.Error("unmarshal memo failed")
		err = errors.New("unmarshal memo failed")
		return
	}

	//bytes in memo data
	memoSize := len(memo.Data) + len(memo.Type)
	if memoSize > memoSpace.MaxMemoSize {
		err = errors.New("full memo size (type + data) must less than 2MB")
		return
	}

	dataBytes, _ := structSerializer.ToMFBytes(memo.MemoData)
	passed, err = memo.Seal.Verify(dataBytes, m.CryptoTools.HashCalculator)

	return
}

func (m *MemoApplication) QueryMemo(hash string) (memo memoSpace.Memo, err error) {
	m.operateLock.RLock()
	defer m.operateLock.RUnlock()

	byteKey, err := hex.DecodeString(hash)
	if err != nil {
		return
	}

	memoData, err := m.kvStorage.Get(byteKey)
	if err != nil {
		return
	}

	if !memoData.Exists {
		err = errors.New("no such memo")
		return
	}

	err = json.Unmarshal(memoData.Data, &memo)
	return
}
