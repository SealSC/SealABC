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
	"errors"
	"github.com/SealSC/SealABC/service/application/traceableStorage/tsData"
)

func (t *TSLedger) GetLocalData(id string) (data tsData.TSData, err error) {
	kvData, err := t.Storage.Get([]byte(id))
	if err != nil {
		return data, errors.New("get prev data error: " + err.Error())
	}

	if !kvData.Exists {
		return data, errors.New("data not found: id [" + id + "]")
	}

	err = data.FromKVStoreItem(kvData.Data)
	if err != nil {
		return data, errors.New("unmarshal data failed: " + err.Error())
	}

	return
}

func (t *TSLedger) VerifyModifyRequest(data tsData.TSData) (err error) {
	//common verify
	_, err = t.commonDataInRequestVerify(data)
	if err != nil {
		return
	}

	//verify prev
	prevData, err := t.GetLocalData(data.PrevOnChainID)
	if err != nil {
		return errors.New("get prev data error: " + err.Error())
	}

	if prevData.NextOnChainID != "" {
		return errors.New("not latest store")
	}

	_, err = data.VerifyPrevSeal(&prevData, t.CryptoTools.HashCalculator)
	return
}

func (t *TSLedger) ExecuteModifyIdentification(data tsData.TSData) (ret interface{}, err error) {
	prevData, err := t.GetLocalData(data.PrevOnChainID)
	if err != nil {
		return
	}

	prevData.NextOnChainID = data.OnChainID
	prevData.SetComplete(t.CryptoTools)

	prevKVItem := prevData.ToKVStoreItem()
	err = t.Storage.Put(prevKVItem)
	if err != nil {
		return
	}

	newStoreKVItem := data.ToKVStoreItem()
	err = t.Storage.Put(newStoreKVItem)
	return
}
