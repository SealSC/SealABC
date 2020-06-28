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
	"SealABC/dataStructure/enum"
	"SealABC/service/application/smartAssets/smartAssetsLedger"
	"SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"errors"
)

type queryParam map[string] string
type queryHandler func(param queryParam) (interface{}, error)

var QueryTypes struct{
	TransactionList enum.Element
	Transaction     enum.Element
	AccountList     enum.Element
	Account         enum.Element
	Contract        enum.Element
	ContractByTx    enum.Element
	ContractList    enum.Element
	ContractCall    enum.Element
	TransferList    enum.Element
}

var QueryParameterFields struct{
	Account  enum.Element
	Contract enum.Element
	TxHash   enum.Element
	Page     enum.Element
}

const rowsPerPage = 20

type commonPagingQueryParam struct {
	queryParam
	rowType       interface{}
	table         string
	condition     string
	conditionArgs []interface{}
}

func (s Storage) commonPagingQuery(param commonPagingQueryParam) (interface{}, error) {
	page := param.queryParam[QueryParameterFields.Page.String()]

	sqlParam := simpleSQLDatabase.SimplePagingQueryParam{
		Count:         rowsPerPage,
		Table:         param.table,
		RowType:       param.rowType,
		Condition:     param.condition,
		ConditionArgs: param.conditionArgs,
	}

	sqlParam.PageFromString(page)

	return s.Driver.SimplePagingQuery(sqlParam)
}

func (s Storage) DoQuery(req smartAssetsLedger.QueryRequest) (interface{}, error) {
	if handler, exists := s.queryHandlers[req.QueryType]; exists {
		return handler(req.Parameter)
	}

	return nil, errors.New("not valid query")
}
