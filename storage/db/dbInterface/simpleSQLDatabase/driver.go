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

package simpleSQLDatabase

import "database/sql"

type IDriver interface {
	Insert(rows IRows, ignoreKey bool) (result sql.Result, err error)
	Replace(rows IRows) (result sql.Result, err error)
	Update(rows IRows, fields []string, condition string, args []interface{}) (result sql.Result, err error)
	Query(rowType interface{}, sql string, args []interface{}) (rows []interface{}, err error)
	SimpleSelect(rowType interface{}, table string, col string, equalVal interface{}) (rows []interface{}, err error)
	RowCount(table string, condition string, args []interface{}) (cnt uint64, err error)
	SimplePagingQuery(param SimplePagingQueryParam) (ret *SimplePagingQueryResult, err error)
}
