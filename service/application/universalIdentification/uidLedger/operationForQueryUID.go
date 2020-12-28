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
	"github.com/SealSC/SealABC/service/application/universalIdentification/uidData"
)

func (u *UIDLedger)QueryUID(queryData uidData.UIDQuery) (ret uidData.QueryResult, err error){
	identification := u.calcIdentification(queryData.PublicKey, queryData.Namespace)
	dataList := u.KVStorage.Traversal([]byte(identification))

	for _, uData := range dataList {
		uid := uidData.UniversalIdentification{}
		_ = json.Unmarshal(uData.Data, &uid)

		ret.UIDList = append(ret.UIDList, uid)
	}

	return
}
