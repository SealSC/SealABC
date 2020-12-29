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

func (u *UIDLedger) verifyUIDKeysUpdate(reqData []byte) (ret interface{}, err error){
	actionData := uidData.UIDUpdateKeys{}
	err = json.Unmarshal(reqData, &actionData)
	if err != nil {
		return nil, errors.New("invalid uid key update data: " + err.Error())
	}

	uData, err := u.KVStorage.Get([]byte(actionData.Identification))
	if err != nil {
		return nil, errors.New("can't get data from db: " + err.Error())
	}

	if !uData.Exists {
		return nil, errors.New("universal identification not exist")
	}

	rawData, _ := structSerializer.ToMFBytes(actionData.UIDUpdateKeysData)
	_, err = actionData.Seal.Verify(rawData, u.CryptoTools.HashCalculator)
	if err != nil {
		return nil, errors.New("invalid signature of append: " + err.Error())
	}

	uid := uidData.UniversalIdentification{}
	_ = json.Unmarshal(uData.Data, &uid)

	for _, newKey := range actionData.NewKeys {
		if newKey.KeyType == uidData.UIDKeyTypes.OracleProof.Int() {
			return nil, errors.New("type of oracle proof key was not supported for now")
		}


		if len(newKey.KeyProof) != 0 {
			return nil, errors.New("self proof was in seal field, key proof field must be empty")
		}

		if newKey.KeyIndex >= len(uid.Keys) {
			return nil, errors.New("key update index out of range")
		}

		uid.Keys[newKey.KeyIndex] = newKey.UIDKey
	}

	newUIDRawData, _ := structSerializer.ToMFBytes(uid.UniversalIdentificationData)

	_, err = actionData.NewUIDSeal.Verify(newUIDRawData, u.CryptoTools.HashCalculator)
	if err != nil {
		return nil, errors.New("invalid signature of new uid: " + err.Error())
	}

	return nil, nil
}

func (u* UIDLedger) updateUIDKeys(reqData []byte) (err error) {
	actionData := uidData.UIDUpdateKeys{}
	_ = json.Unmarshal(reqData, &actionData)

	uData, _ := u.KVStorage.Get([]byte(actionData.Identification))
	uid := uidData.UniversalIdentification{}
	_ = json.Unmarshal(uData.Data, &uid)

	for _, newKey := range actionData.NewKeys {
		uid.Keys[newKey.KeyIndex] = newKey.UIDKey
	}
	uid.Seal = actionData.NewUIDSeal

	newData, _ := json.Marshal(&uid)

	_ = u.KVStorage.Put(kvDatabase.KVItem {
		Key:    []byte(uid.Identification),
		Data:   newData,
		Exists: true,
	})
	return
}
