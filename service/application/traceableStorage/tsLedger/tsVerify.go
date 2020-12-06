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

import "errors"

func (t *TSLedger) commonDataInRequestVerify(data TSData) (passed bool, err error) {
	//verify chain id
	_, err = data.Verify(t.CryptoTools.HashCalculator)
	if err != nil {
		return
	}

	//verify id duplicate
	localData, err := t.Storage.Get([]byte(data.OnChainID))
	if err != nil {
		return
	}

	if localData.Exists {
		return false, errors.New("duplicate key")
	}

	//verify complete pk
	if data.CompleteKey == "" {
		return false, errors.New("complete key must not empty")
	}

	//verify next chain id
	if data.NextOnChainID != "" {
		return false, errors.New("next chain id can only be calc by system")
	}

	if !data.CompleteSeal.IsPureEmpty() {
		return false, errors.New("next chain seal in request must empty")
	}

	return true, nil
}

