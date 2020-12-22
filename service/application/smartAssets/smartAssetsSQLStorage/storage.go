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
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

type Storage struct {
	queryHandlers map[string] queryHandler
	Driver simpleSQLDatabase.IDriver
}

func Load() {
	enum.SimpleBuild(&QueryTypes)
	enum.SimpleBuild(&QueryParameterFields)
}

func NewStorage(sqlDriver simpleSQLDatabase.IDriver) (s *Storage) {
	s = &Storage{
		Driver:        sqlDriver,
	}

	s.queryHandlers = map[string] queryHandler {
		QueryTypes.TransactionList.String(): s.queryTransactionList,
		QueryTypes.Transaction.String():     s.queryTransactionByHash,
		QueryTypes.AccountList.String():     s.queryAccountList,
		QueryTypes.Account.String():         s.queryAccount,
		QueryTypes.Contract.String():        s.queryContractByAddress,
		QueryTypes.ContractByTx.String():    s.queryContractByTx,
		QueryTypes.ContractList.String():    s.queryContractList,
		QueryTypes.ContractCall.String():    s.queryContractCallByHash,
		QueryTypes.TransferList.String():    s.queryTransferList,
	}

	return
}
