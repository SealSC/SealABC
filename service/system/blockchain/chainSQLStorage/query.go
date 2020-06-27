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

package chainSQLStorage

import (
    "SealABC/metadata/httpJSONResult/rowsWithCount"
    "SealABC/service/system/blockchain/chainTables"
    "errors"
)

const rowsPerPage = 20

func (s *Storage) GetBlockList(start uint64) (blocks []chainTables.BlockListRow, err error) {
    result, err := s.Driver.Query(chainTables.BlockListRow{},
        "select * from `t_block_list` where `c_height`>=? order by `c_height` asc limit 0,?",
        []interface{}{start, rowsPerPage})

    if err != nil {
        return
    }

    for _, r := range result {
        newBlk := r.(chainTables.BlockListRow)
        newBlk.Payload = ""
        blocks = append(blocks, newBlk)
    }

    return
}

func (s *Storage) GetBlock(height uint64) (blk chainTables.BlockListRow, err error) {
    result, err := s.Driver.Query(chainTables.BlockListRow{},
        "select * from `t_block_list` where `c_height`=? ",
        []interface{}{height})

    if len(result) != 1 {
        err = errors.New("no such block")
        return
    }

    blk, _  = result[0].(chainTables.BlockListRow)
    return
}

func (s *Storage) GetRequestList(page uint64) (ret rowsWithCount.Entity, err error) {
    startId := page * rowsPerPage

    table := chainTables.Requests.Name()
    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_id`>=? order by `c_id` desc limit 0,?"

    row := chainTables.RequestRow{}
    rows, err := s.Driver.Query(row, pSQL, []interface{} {
        startId,
        rowsPerPage,
    })

    if err != nil {
        return
    }

    count, err := s.Driver.RowCount(table, "", nil)
    if err != nil {
        return
    }

    list := rowsWithCount.Entity {
        Rows:  rows,
        Total: count,
    }

    return list, err
}

func (s *Storage) GetBlockByHash(hash string) (blk chainTables.BlockListRow, err error) {
    rows, err := s.Driver.SimpleSelect (
        chainTables.BlockListRow{},
        chainTables.BlockList.Name(),
        `c_hash`,
        hash,
    )

    if err != nil {
        return
    }

    if len(rows) != 1 {
        err = errors.New("no such block")
        return
    }

    blk = rows[0].(chainTables.BlockListRow)
    return
}

func (s *Storage) GetRequestByHash(hash string) (req chainTables.RequestRow, err error) {
    table := chainTables.Requests.Name()
    pSQL := "select * from `" + table + "` where `c_hash`=?"

    rows, err := s.Driver.Query(chainTables.RequestRow{}, pSQL, []interface{}{hash})
    if err != nil {
        return
    }

    if len(rows) == 0 {
        err = errors.New("no such transaction")
        return
    }

    req = rows[0].(chainTables.RequestRow)
    return
}

func (s *Storage) GetRequestByHeight(height string) (ret rowsWithCount.Entity, err error) {

    table := chainTables.Requests.Name()
    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_height`=? order by `c_id` desc"

    row := chainTables.RequestRow{}
    rows, err := s.Driver.Query(row, pSQL, []interface{} {height})

    if err != nil {
        return
    }

    list := rowsWithCount.Entity {
        Rows:  rows,
        Total: uint64(len(rows)),
    }

    return list, err
}

func (s *Storage) GetRequestByApplicationAndAction(app string, act string, page uint64) (ret rowsWithCount.Entity, err error) {

    table := chainTables.Requests.Name()
    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_application`=? "

    if act != "*" {
        pSQL += " and `c_action`=? "
    }

    pSQL += " order by `c_id` desc limit ?,?"

    start := page * rowsPerPage

    sqlParam := []interface{} {app}

    if act != "*" {
        sqlParam = append(sqlParam, act)
    }
    sqlParam = append(sqlParam, start, rowsPerPage)

    row := chainTables.RequestRow{}
    rows, err := s.Driver.Query(row, pSQL, sqlParam)

    if err != nil {
        return
    }

    list := rowsWithCount.Entity {
        Rows:  rows,
        Total: uint64(len(rows)),
    }

    return list, err
}
