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

package basicAssetsSQLStorage

import (
    "github.com/SealSC/SealABC/metadata/httpJSONResult/rowsWithCount"
    "github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsSQLTables"
    "errors"
)

const rowsPerPage = 20

func (s *Storage) GetAssetsList(p []string) (ret interface{}, err error) {
    page, err := pageFromParam(p)
    if err != nil {
        return
    }

    table := basicAssetsSQLTables.AssetsList.Name()
    rowType := basicAssetsSQLTables.AssetsListRow{}
    start := rowsPerPage * page

    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_id` >=?  order by `c_id` desc limit 0,?"

    rows, err := s.Driver.Query(rowType, pSQL,
        []interface{}{
            start,
            rowsPerPage,
        })

    if err != nil {
        return
    }

    count, err := s.Driver.RowCount(table, "", nil)
    if err != nil {
        return
    }

    result := rowsWithCount.Entity {
        Rows:  rows,
        Total: count,
    }

    return result, err
}

func (s *Storage) GetTransferList(p []string) (ret interface{}, err error) {
    page, err := pageFromParam(p)
    if err != nil {
        return
    }

    rowType := basicAssetsSQLTables.TransfersRow{}
    table := basicAssetsSQLTables.Transfers.Name()
    count, err := s.Driver.RowCount(table, "", nil)
    if err != nil {
        return
    }

    start := count - (page * rowsPerPage)
    queryData := []interface{} {
        start,
        rowsPerPage,
    }

    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_id`<=? order by `c_id` desc limit 0,?"

    rows, err := s.Driver.Query(rowType, pSQL, queryData)
    if err != nil {
        return
    }

    result := rowsWithCount.Entity {
        Rows: rows,
        Total: count,
    }

    return result, err
}

func (s *Storage) GetTransfer(p []string) (ret interface{}, err error) {
    txHash, err := hashFromParam(p)
    if err != nil {
        return
    }

    rowType := basicAssetsSQLTables.TransfersRow{}
    table := basicAssetsSQLTables.Transfers.Name()

    rows, err := s.Driver.SimpleSelect(rowType, table, `c_tx_hash`, txHash)
    if err != nil {
        return
    }

    if len(rows) == 0 {
        err = errors.New("no such transfer")
        return
    }

    ret = rows[0]
    return
}

func (s *Storage) GetTransfersUnderAssets(p []string) (ret interface{}, err error) {
    page, assets, err := pageAndHashFromParam(p)
    if err != nil {
        return
    }

    row := basicAssetsSQLTables.TransfersRow{}
    table := basicAssetsSQLTables.Transfers.Name()
    count, err := s.Driver.RowCount(table, "where `c_assets_hash`=?", []interface{}{assets})
    if err != nil {
        return
    }

    start := page * rowsPerPage
    queryData := []interface{} {
        assets,
        start,
        rowsPerPage,
    }

    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_assets_hash`=? order by `c_id` desc limit ?,?"

    rows, err := s.Driver.Query(row, pSQL, queryData)
    if err != nil {
        return
    }

    result := rowsWithCount.Entity {
        Rows: rows,
        Total: count,
    }

    return result, err
}

func (s *Storage) GetAssets(p []string) (ret interface{}, err error)  {
    assetsHash, err := hashFromParam(p)
    if err != nil {
        return
    }

    row := basicAssetsSQLTables.AssetsListRow{}
    table :=  basicAssetsSQLTables.AssetsList.Name()

    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_meta_hash`=?"

    assets, err := s.Driver.Query(row, pSQL, []interface{}{assetsHash})
    if err != nil {
        return
    }

    if len(assets) < 1 {
        err = errors.New("no such assets")
        return
    }

    return assets[0], err
}

func (s *Storage) GetAddressesList(p []string) (ret interface{}, err error) {
    page, err := pageFromParam(p)
    if err != nil {
        return
    }

    table := basicAssetsSQLTables.AddressList.Name()
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

    row := basicAssetsSQLTables.AddressListRow{}
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

func (s *Storage) GetBalancesUnderAssetsList(p []string) (ret interface{}, err error) {
    page, assets, err := pageAndHashFromParam(p)
    if err != nil {
        return
    }

    table := basicAssetsSQLTables.Balance.Name()
    rowType := basicAssetsSQLTables.BalanceRow{}
    count, err := s.Driver.RowCount(table, "where `c_assets`=?", []interface{}{assets})
    if err != nil {
        return
    }

    startPage := page * rowsPerPage
    pSQL := "select * from " +
        "`" + table + "`" +
        " where `c_assets`=? order by `c_id` desc limit ?,?"

    rows, err := s.Driver.Query(rowType, pSQL, []interface{} {
        assets,
        startPage,
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

func (s *Storage) GetAddressActionRecord(p []string) (ret interface{}, err error) {
    page, address, err := pageAndHashFromParam(p)
    if err != nil {
        return
    }

    table := basicAssetsSQLTables.AddressRecord.Name()
    rowType := basicAssetsSQLTables.AddressRecordRow{}

    offsetStart := page * rowsPerPage

    pSQL := "select * from " +
        "`" + table +  "`" +
        " where `c_address`=? order by `c_id` desc limit ?,?"

    rows, err := s.Driver.Query(rowType, pSQL, []interface{} {
        address,
        offsetStart,
        rowsPerPage,
    })

    if err != nil {
        return
    }

    count, err := s.Driver.RowCount(table, "where `c_address`=?", []interface{}{
        address,
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

func (s *Storage) GetAddressBalance(p []string) (ret interface{}, err error) {
    page, addr, err := pageAndHashFromParam(p)
    if err != nil {
        return
    }

    table := basicAssetsSQLTables.Balance.Name()
    rowType := basicAssetsSQLTables.BalanceRow{}

    count, err := s.Driver.RowCount(table, "where `c_address`=?", []interface{}{
        addr,
    })
    if err != nil {
        return
    }

    offsetStart := page * rowsPerPage
    pSQL := "select * from `" + table +"` where `c_address`=? order by `c_id` asc limit ?,?"

    rows, err := s.Driver.Query(rowType, pSQL, []interface{}{
        addr,
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


func (s *Storage) GetSellingHistory(p []string) (ret interface{}, err error) {
    table := basicAssetsSQLTables.SellingList.Name()
    rowType := basicAssetsSQLTables.SellingListRow{}

    count, err := s.Driver.RowCount(table, "", []interface{}{})
    if err != nil {
        return
    }

    pSQL := "select * from `" + table +"` order by `c_id` desc"

    rows, err := s.Driver.Query(rowType, pSQL, []interface{}{})

    if err != nil {
        return
    }

    list := rowsWithCount.Entity {
        Rows:  rows,
        Total: count,
    }

    return list, err
}
