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
	"errors"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/service/application/universalIdentification/uidData"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

func Load()  {
	enum.SimpleBuild(&uidData.UIDKeyTypes)
	enum.SimpleBuild(&uidData.UIDActionTypes)
}

func NewLedger(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver) (ledger UIDLedger) {
	ledger.KVStorage = kvDriver

	ledger.validators = map[string]actionValidator {
		uidData.UIDActionTypes.Create.String(): ledger.verifyUIDCreation,
		uidData.UIDActionTypes.Append.String(): ledger.verifyUIDKeysAppend,
		uidData.UIDActionTypes.Update.String(): ledger.verifyUIDKeysUpdate,
	}

	ledger.executors = map[string]actionExecutor {
		uidData.UIDActionTypes.Create.String(): ledger.createUID,
		uidData.UIDActionTypes.Append.String(): ledger.appendUIDKeys,
		uidData.UIDActionTypes.Update.String(): ledger.updateUIDKeys,
	}

	return
}

type actionValidator func(actData []byte) (ret interface{}, err error)
type actionExecutor func(actData []byte) (err error)

type UIDLedger struct {
	validators map[string] actionValidator
	executors map[string] actionExecutor

	CryptoTools crypto.Tools
	KVStorage   kvDatabase.IDriver
}

func (u *UIDLedger) VerifyAction(action string, data []byte) (ret interface{}, err error){
	if validate, exists := u.validators[action]; exists {
		return validate(data)
	}

	return nil, errors.New("action not supported")
}

func (u *UIDLedger) ExecuteAction(action string, data []byte) (err error)  {
	if executor, exists := u.executors[action]; exists {
		return executor(data)
	}

	return errors.New("action not supported")
}

func (u *UIDLedger) calcIdentification(pubKey []byte, namespace string) string {
	rawID := string(pubKey) + namespace
	return u.CryptoTools.HashCalculator.SumHex([]byte(rawID))
}
