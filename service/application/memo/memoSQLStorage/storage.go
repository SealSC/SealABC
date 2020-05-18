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

package memoSQLStorage

import (
    "SealABC/dataStructure/enum"
    "SealABC/service/application/memo/memoSpace"
    "SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "errors"
)

var QueryTypes struct{
    MemoList    enum.Element
    MemoByHash  enum.Element
    MemoByType  enum.Element

    AddressList         enum.Element
    MemoUnderAddress    enum.Element
    Statistics          enum.Element
}

type queryHandler func([]string) (interface{}, error)
type Storage struct {
    queryHandlers map[string] queryHandler
    Driver simpleSQLDatabase.IDriver
}

func Load()  {
    enum.SimpleBuild(&QueryTypes)
}

func NewStorage(sqlDriver simpleSQLDatabase.IDriver) (s *Storage) {
    s = &Storage{
        Driver:        sqlDriver,
    }

    s.queryHandlers = map[string] queryHandler {
        QueryTypes.MemoList.String(): s.GetMemoList,
        QueryTypes.MemoByHash.String(): s.GetMemoByHash,
        QueryTypes.MemoByType.String(): s.GetMemoByType,
        QueryTypes.AddressList.String(): s.GetAddressList,
        QueryTypes.MemoUnderAddress.String(): s.GetMemoUnderAddress,
        QueryTypes.Statistics.String(): s.GetStatistics,
    }
    return
}

func (s *Storage) DoQuery(queryReq memoSpace.QueryRequest) (result interface{}, err error) {
    if handler, exists := s.queryHandlers[queryReq.QueryType]; !exists {
        err = errors.New("no such query handler: " + queryReq.QueryType)
        return
    } else {
        return handler(queryReq.Parameter)
    }
}
