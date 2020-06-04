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

package basicAssetsSQLTables

import (
    "SealABC/common"
    "SealABC/dataStructure/enum"
    "SealABC/service/application/basicAssets/basicAssetsLedger"
    "SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "encoding/hex"
    "fmt"
    "time"
)

type BalanceTable struct {
    ID              enum.Element `col:"c_id" ignoreInsert:"true"`
    LastHeight      enum.Element `col:"c_last_height"`
    Address         enum.Element `col:"c_address"`
    Assets          enum.Element `col:"c_assets"`
    AssetsName      enum.Element `col:"c_assets_name"`
    AssetsSymbol    enum.Element `col:"c_assets_symbol"`
    Amount          enum.Element `col:"c_amount"`
    Time            enum.Element `col:"c_time"`

    simpleSQLDatabase.BasicTable
}

var Balance BalanceTable

func (b BalanceTable) NewRows() interface{} {
    return simpleSQLDatabase.NewRowsInstance(BalanceRows{})
}

func (b BalanceTable) Name() (name string) {
    return "t_basic_assets_balance"
}

func (b *BalanceTable) load() {
    enum.SimpleBuild(b)
    b.Instance = *b
}

type BalanceRow struct {
    ID              string
    LastHeight      string
    Address         string
    Assets          string
    AssetsName      string
    AssetsSymbol    string
    Amount          string
    Time            string
}

type BalanceRows struct {
    simpleSQLDatabase.BasicRows
}

func (b *BalanceRows) InsertBalances(height uint64, tm int64, balanceList []basicAssetsLedger.Balance)  {
    timestamp := time.Unix(tm, 0)
    for _, balance := range balanceList {
        newAddressRow := BalanceRow{
            LastHeight: fmt.Sprintf("%d", height),
            Address: hex.EncodeToString(balance.Address),
            Assets: hex.EncodeToString(balance.Assets.MetaSeal.Hash),
            AssetsName: balance.Assets.Name,
            AssetsSymbol: balance.Assets.Symbol,
            Amount: fmt.Sprintf("%d", balance.Amount),
            Time:    timestamp.Format(common.BASIC_TIME_FORMAT),
        }

        b.Rows = append(b.Rows, newAddressRow)
    }
}

func (b *BalanceRows) Table() simpleSQLDatabase.ITable {
    return &Balance
}
