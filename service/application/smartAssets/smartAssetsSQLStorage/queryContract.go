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
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsSQLTables"
	"errors"
)

var contractRowType = smartAssetsSQLTables.ContractRow{}
var contractTableName = smartAssetsSQLTables.Contract.Name()

func (s Storage) queryContractList(param queryParam) (interface{}, error) {
	account := param[QueryParameterFields.Account.String()]

	commonParam := commonPagingQueryParam{
		queryParam:    param,
		rowType:       contractRowType,
		table:         contractTableName,
	}

	if account != "" {
		commonParam.condition = " where `c_creator`=? "
		commonParam.conditionArgs = []interface{}{account}
	}

	return s.commonPagingQuery(commonParam)
}

func (s Storage) queryContractByTx(param queryParam) (interface{}, error){
	txHash := param[QueryParameterFields.TxHash.String()]
	if txHash == "" {
		return nil, errors.New("invalid parameters")
	}

	return s.Driver.SimpleSelect(contractRowType, contractTableName, `c_tx_hash`, txHash)
}

func (s Storage) queryContractByAddress(param queryParam) (interface{}, error){
	address := param[QueryParameterFields.Contract.String()]
	if address == "" {
		return nil, errors.New("invalid parameters")
	}

	return s.Driver.SimpleSelect(contractRowType, contractTableName, `c_contract_address`, address)
}
