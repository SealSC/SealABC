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

package uidLedger

import (
	"encoding/json"
	"errors"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/service/application/universalIdentification/uidData"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
)

func (u *UIDLedger) verifyUIDCreation(reqData []byte) (ret interface{}, err error) {
	actionData := uidData.UIDCreation{}
	err = json.Unmarshal(reqData, &actionData)
	if err != nil {
		return nil, errors.New("invalid uid creation data: " + err.Error())
	}

	uid := actionData.UID
	newHashedID := u.calcIdentification(uid.Seal.SignerPublicKey, uid.Namespace)

	if newHashedID != uid.Identification {
		return nil, errors.New("identification not equal")
	}

	data, err := u.KVStorage.Get([]byte(newHashedID))
	if err != nil {
		return nil, errors.New("get data from db failed: " + err.Error())
	}

	if data.Exists {
		return nil, errors.New("uid already exist")
	}

	rawTxData, _ := structSerializer.ToMFBytes(uid)
	_, err = actionData.Seal.Verify(rawTxData, u.CryptoTools.HashCalculator)
	if err != nil {
		return nil, errors.New("invalid signature of transaction: " + err.Error())
	}

	for _, key := range uid.Keys {
		if key.KeyType == uidData.UIDKeyTypes.OracleProof.Int() {
			return nil, errors.New("type of oracle proof key was not supported for now")
		}

		if len(key.KeyProof) != 0 {
			return nil, errors.New("self proof was in seal field, key proof field must be empty")
		}
	}

	rawUIDData, _ := structSerializer.ToMFBytes(uid.UniversalIdentificationData)
	_, err = uid.Seal.Verify(rawUIDData, u.CryptoTools.HashCalculator)
	if err != nil {
		return nil, errors.New("invalid signature of uid data: " + err.Error())
	}

	return nil, nil
}

func (u *UIDLedger) createUID(reqData []byte) (err error) {
	actionData := uidData.UIDCreation{}
	_ = json.Unmarshal(reqData, &actionData)

	dataToStore, _ := json.Marshal(actionData.UID)

	err = u.KVStorage.Put(kvDatabase.KVItem{
		Key:    []byte(actionData.UID.Identification),
		Data:   dataToStore,
		Exists: true,
	})

	return
}
