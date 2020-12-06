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

package tsLedger

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/crypto"
	"SealABC/crypto/hashes"
	"SealABC/dataStructure/enum"
	"SealABC/metadata/seal"
	"SealABC/storage/db/dbInterface/kvDatabase"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type TSMetaData struct {
	Namespace   string
	ExternalID  string
	RawData     string
	CompleteKey string
}

type TSCompleteData struct {
	TSMetaData

	OnChainID     string
	PrevOnChainID string
	NextOnChainID string
}

type TSData struct {
	TSCompleteData

	Seal         seal.Entity
	PrevSeal     seal.Entity
	CompleteSeal seal.Entity
}

func (i *TSData) GetMetaDataBytes() (data []byte) {
	var buff bytes.Buffer
	buff.Write([]byte(i.Namespace))
	buff.Write([]byte(i.ExternalID))
	buff.Write([]byte(i.RawData))
	buff.Write([]byte(i.CompleteKey))
	return buff.Bytes()
}

func (i *TSData) GetMetaDataBytesWithSealBytes() (data []byte) {
	metaBytes := i.GetMetaDataBytes()
	sealBytes, _ := structSerializer.ToMFBytes(i.Seal)

	var buff bytes.Buffer
	buff.Write(metaBytes)
	buff.Write(sealBytes)
	return buff.Bytes()
}

func (i *TSData) GetCompleteBytes() (data []byte, err error) {
	return structSerializer.ToMFBytes(i.TSCompleteData)
}

func (i *TSData) ToMFBData() (data []byte, err error) {
	return structSerializer.ToMFBytes(i)
}

func (i *TSData) CalcChainID(hashCalc hashes.IHashCalculator) (id string, metaBytes []byte) {
	metaBytes = i.GetMetaDataBytes()
	idHash := hashCalc.Sum(metaBytes)

	id = hex.EncodeToString(idHash)
	return
}

func (i *TSData) SetComplete(tools crypto.Tools) {
	data, _ := i.GetCompleteBytes()
	var buff bytes.Buffer
	completeKeyAsPrivateKey := tools.HashCalculator.Sum([]byte(i.CompleteKey))
	buff.Write(completeKeyAsPrivateKey)
	buff.Write(completeKeyAsPrivateKey)
	_ = i.CompleteSeal.Sign(data, tools, buff.Bytes())
}

func (i *TSData) Verify(hashCalc hashes.IHashCalculator) (passed bool, err error) {
	id, metaBytes := i.CalcChainID(hashCalc)
	if id != i.OnChainID {
		return false, errors.New("invalid chain id: " + id + " != " + i.OnChainID)
	}

	passed, err = i.Seal.Verify(metaBytes, hashCalc)
	if !passed {
		return passed, errors.New("invalid meta seal: " + err.Error())
	}

	if i.NextOnChainID != "" {
		completeBytes, _ := i.GetCompleteBytes()
		passed, err = i.CompleteSeal.Verify(completeBytes, hashCalc)
		if !passed {
			return passed, errors.New("invalid complete seal")
		}
	}

	return true, nil
}

func (i *TSData) VerifyPrevSeal(prevData *TSData, hashCalc hashes.IHashCalculator) (passed bool, err error) {
	metaWithSeal := i.GetMetaDataBytesWithSealBytes()
	passed, err = i.PrevSeal.Verify(metaWithSeal, hashCalc)
	if !passed {
		return passed, errors.New("invalid prev seal: " + err.Error())
	}

	if !bytes.Equal(prevData.Seal.SignerPublicKey, i.PrevSeal.SignerPublicKey) {
		return passed, errors.New("prev seal not equal")
	}

	return true, nil
}

func (i *TSData) ToKVStoreItem() (item kvDatabase.KVItem) {
	data, _ := json.Marshal(i)
	item = kvDatabase.KVItem{
		Key:    []byte(i.OnChainID),
		Data:   data,
		Exists: true,
	}
	return
}

func (i *TSData) FromKVStoreItem(data []byte) (err error) {
	err = json.Unmarshal(data, i)
	return
}

var RequestTypes struct {
	Store   enum.Element
	Modify  enum.Element
}

type TSServiceRequest struct {
	ReqType string
	Data    TSData
}
