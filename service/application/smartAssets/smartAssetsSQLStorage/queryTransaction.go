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

var txRowType = smartAssetsSQLTables.TransactionRow{}
var txTableName = smartAssetsSQLTables.Transaction.Name()

func (s Storage) queryTransactionList(param queryParam) (interface{}, error) {
	account := param[QueryParameterFields.Account.String()]

	commonParam := commonPagingQueryParam{
		queryParam: param,
		rowType:    txRowType,
		table:      txTableName,
	}

	if account != "" {
		commonParam.condition = " where `c_from`=? or `c_to`=? "
		commonParam.conditionArgs = []interface{}{account, account}
	}

	return s.commonPagingQuery(commonParam)
}

func (s Storage) queryTransactionByHash(param queryParam) (interface{}, error) {
	txHash := param[QueryParameterFields.TxHash.String()]
	if txHash == "" {
		return nil, errors.New("invalid parameters")
	}

	return s.Driver.SimpleSelect(txRowType, txTableName, `c_hash`, txHash)
}
