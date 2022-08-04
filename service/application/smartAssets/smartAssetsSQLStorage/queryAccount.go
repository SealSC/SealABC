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

package smartAssetsSQLStorage

import (
	"errors"
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsSQLTables"
)

var addressListRowType = smartAssetsSQLTables.AddressListRow{}
var addressListTableName = smartAssetsSQLTables.AddressList.Name()

func (s Storage) queryAccountList(param queryParam) (interface{}, error) {
	commonParam := commonPagingQueryParam{
		queryParam: param,
		rowType:    addressListRowType,
		table:      addressListTableName,
	}

	return s.commonPagingQuery(commonParam)
}

func (s Storage) queryAccount(param queryParam) (interface{}, error) {
	account := param[QueryParameterFields.Account.String()]
	if account == "" {
		return nil, errors.New("invalid parameters")
	}

	return s.Driver.SimpleSelect(addressListRowType, addressListTableName, `c_address`, account)
}
