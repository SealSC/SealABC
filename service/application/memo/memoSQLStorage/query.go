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
    "SealABC/metadata/httpJSONResult/rowsWithCount"
    "SealABC/service/application/memo/memoTables"
    "errors"
)

const rowsPerPage = 20

func (s *Storage) GetMemoList(p []string) (ret interface{}, err error) {
    page, err := pageFromParam(p)
    if err != nil {
        return
    }

    table := memoTables.MemoList.Name()
    rowType := memoTables.MemoListRow{}

    count, err := s.Driver.RowCount(table, "", nil)
    if err != nil {
        return
    }

    offsetStart := rowsPerPage * page
    pSQL := "select * from " +
        "`" + table + "`" +
        " order by `c_id` desc limit ?,?"

    rows, err := s.Driver.Query(rowType, pSQL,
        []interface{}{
            offsetStart,
            rowsPerPage,
        })

    if err != nil {
        return
    }

    result := rowsWithCount.Entity {
        Rows:  rows,
        Total: count,
    }

    return result, err
}


func (s *Storage) GetAddressList(p []string) (ret interface{}, err error) {
    page, err := pageFromParam(p)
    if err != nil {
        return
    }

    table := memoTables.AddressList.Name()
    offsetStart := page * rowsPerPage

    count, err := s.Driver.RowCount(table, "", nil)
    if err != nil {
        return
    }

    if offsetStart > count {
        return rowsWithCount.Entity {
            Rows:  nil,
            Total: count,
        }, nil
    }

    pSQL := "select * from " +
        "`" + table + "`" +
        " order by `c_id` desc limit ?,?"

    row := memoTables.AddressListRow{}
    rows, err := s.Driver.Query(row, pSQL, []interface{} {
        offsetStart,
        rowsPerPage,
    })
    if err != nil {
        return
    }

    list := rowsWithCount.Entity {
        Rows:  rows,
        Total: count,
    }

    return list, err
}

func (s *Storage) GetMemoByHash(p []string) (memoRow interface{}, err error) {
    hash, err := hashFromParam(p)
    if err != nil {
        return
    }

    rowType := memoTables.MemoListRow {}
    table := memoTables.MemoList.Name()
    rows, err := s.Driver.SimpleSelect(rowType, table, `c_hash`, hash)
    if err != nil {
        return
    }

    if len(rows) != 1 {
        err = errors.New("no such memo")
        return
    }

    memoRow = rows[0]
    return
}

func (s *Storage) GetMemoByType(p []string) (memoRows interface{}, err error) {
    page, memoType, err := pageAndMemoTypeFromParam(p)
    if err != nil {
        return
    }

    table := memoTables.MemoList.Name()
    start := page * rowsPerPage
    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_type`=? order by `c_id` desc limit ?,?"

    rowType := memoTables.MemoListRow{}
    rows, err := s.Driver.Query(rowType, pSQL, []interface{}{
        memoType,
        start,
        rowsPerPage,
    })

    if err != nil {
        return
    }

    count, err := s.Driver.RowCount(table, "where `c_type`=?", []interface{}{memoType})
    if err != nil {
        return
    }

    memoRows = rowsWithCount.Entity {
        Rows: rows,
        Total: count,
    }


    return
}

func (s *Storage) GetMemoUnderAddress(p []string) (memoRows interface{}, err error) {
    page, address, err := pageAndMemoTypeFromParam(p)
    if err != nil {
        return
    }

    table := memoTables.MemoList.Name()
    start := page * rowsPerPage
    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_recorder`=? order by `c_id` desc limit ?,?"

    rowType := memoTables.MemoListRow{}
    rows, err := s.Driver.Query(rowType, pSQL, []interface{}{
        address,
        start,
        rowsPerPage,
    })

    if err != nil {
        return
    }

    count, err := s.Driver.RowCount(table, "where `c_recorder`=?", []interface{}{address})
    if err != nil {
        return
    }

    memoRows = rowsWithCount.Entity {
        Rows: rows,
        Total: count,
    }

    return
}
