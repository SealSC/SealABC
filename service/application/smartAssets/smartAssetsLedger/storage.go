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

package smartAssetsLedger

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/dataStructure/enum"
)

const commonKeyLength = 32

type prefixEl struct {
	enum.Element
}

func (p prefixEl) buildKey(baseKey []byte, extra ...[]byte) []byte {
	result := append([]byte(p.String()), baseKey...)
	for _, k := range extra {
		result = append(result, k...)
	}
	return result
}

var StoragePrefixes struct {
	Assets       prefixEl
	Transaction  prefixEl
	ContractData prefixEl
	ContractCode prefixEl
	Balance      prefixEl
}

func (l Ledger) getTxFromStorage(hash []byte) (tx *Transaction, exists bool, err error) {
	key := StoragePrefixes.Transaction.buildKey(hash)

	txJson, err := l.Storage.Get(key)
	if err != nil {
		return
	}

	exists = txJson.Exists
	if exists {
		tx = &Transaction{}
		err = structSerializer.FromMFBytes(txJson.Data, tx)
	}

	return
}
