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

func (t *TSLedger)VerifyStoreRequest(data tsData.TSData) (err error) {
	//common verify
	_, err = t.commonDataInRequestVerify(data)
	if err != nil {
		return
	}

	//verify prev & next id
	if data.PrevOnChainID != "" {
		return errors.New("new data must has no prev or next chain id")
	}

	if !data.PrevSeal.IsPureEmpty() {
		return errors.New("new data must has no prev seal ")
	}

	return nil
}

func (t *TSLedger)ExecuteStoreIdentification(data tsData.TSData) (ret interface{}, err error) {
	kvData := data.ToKVStoreItem()
	err = t.Storage.Put(kvData)
	return
}
