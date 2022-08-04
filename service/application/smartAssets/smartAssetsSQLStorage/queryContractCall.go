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
	"strings"
)

var contractCallRowType = smartAssetsSQLTables.ContractCallRow{}
var contractCallTableName = smartAssetsSQLTables.ContractCall.Name()

func buildContractCallQueryCondition(account string, contract string) (string, []interface{}) {
	var condition []string
	var args []interface{}
	if account != "" {
		condition = append(condition, " `c_caller`=? ")
		args = append(args, account)
	}

	if contract != "" {
		condition = append(condition, " `c_contract_address`=? ")
		args = append(args, contract)
	}

	conditionStr := strings.Join(condition, " and ")

	return " where " + conditionStr, args
}

func (s Storage) queryContractCallList(param queryParam) (interface{}, error) {
	account := param[QueryParameterFields.Account.String()]
	contract := param[QueryParameterFields.Contract.String()]

	commonParam := commonPagingQueryParam{
		queryParam: param,
		rowType:    contractCallRowType,
		table:      contractCallTableName,
	}

	if account != "" || contract != "" {
		condition, args := buildContractCallQueryCondition(account, contract)

		commonParam.condition = condition
		commonParam.conditionArgs = args
	}

	return s.commonPagingQuery(commonParam)
}

func (s Storage) queryContractCallByHash(param queryParam) (interface{}, error) {
	txHash := param[QueryParameterFields.TxHash.String()]
	if txHash == "" {
		return nil, errors.New("invalid parameters")
	}

	return s.Driver.SimpleSelect(contractCallRowType, contractCallTableName, `c_tx_hash`, txHash)
}
