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

package simpleMysql

import (
    "SealABC/dataStructure/enum"
    "SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "database/sql"
    "errors"
    _ "github.com/go-sql-driver/mysql"
    "strings"
)

var Charsets struct{
    UTF8 enum.Element
}

type Config struct {
    User            string
    Password        string
    DBName          string
    Charset         string
    MaxConnection   int
}

type simpleMySQLDriver struct {
    db *sql.DB
}

func (s *simpleMySQLDriver) exec(pSQL string, data []interface{}) (result sql.Result, err error)  {
    stmt, err := s.db.Prepare(pSQL)
    if err != nil {
        return
    }
    defer func() {
        err = stmt.Close()
    }()

    result, err = stmt.Exec(data...)
    return
}

func (s *simpleMySQLDriver) getInsertContentSQL(sqlPrefix string, rows simpleSQLDatabase.IRows) (pSQL string, data []interface{})  {
    pSQL = sqlPrefix

    rowCount := rows.Count()
    table := rows.Table()

    columns := table.ColumnsForInsert()
    tName := table.Name()

    qm := make([]string, len(columns))
    for i := 0; i < len(columns); i++ {
        qm[i] = "?"
    }
    pSQL += " `" + tName + "` (`" + strings.Join(columns[:], "`,`") + "`)"

    var values []string
    for i := 0; i < rowCount; i++ {
        values = append(values, "("+strings.Join(qm[:], ",")+")")
    }

    pSQL += " values " + strings.Join(values, ",")
    data =  rows.DataForInsert()
    return
}

func (s *simpleMySQLDriver) insert(sqlPrefix string, rows simpleSQLDatabase.IRows) (result sql.Result, err error) {
    pSQL, data := s.getInsertContentSQL(sqlPrefix, rows)
    result, err = s.exec(pSQL , data)

    return
}

func (s *simpleMySQLDriver) Insert(rows simpleSQLDatabase.IRows, ignoreKey bool) (result sql.Result, err error) {
    var pSQL string
    if ignoreKey {
        pSQL = "insert ignore into "
    } else {
        pSQL = "insert into "
    }

    result, err = s.insert(pSQL , rows)
    return
}

func (s *simpleMySQLDriver) Replace(rows simpleSQLDatabase.IRows) (result sql.Result, err error) {
    pSQL := "replace into "
    result, err = s.insert(pSQL , rows)
    return
}

func (s *simpleMySQLDriver) Update(rows simpleSQLDatabase.IRows, columns []string, condition string, args []interface{}) (result sql.Result, err error) {
    pSQL := "update ? set "
    var data []interface{}

    columnCount := len(columns)
    table := rows.Table()

    tName := table.Name()

    data = append(data, tName)

    dataSegment := make([]string, columnCount)
    uData := rows.Data(columns)

    for i:=0; i<columnCount; i++ {
        dataSegment[i] = "?=?"
        data = append(data, columns[i], uData[i])
    }

    pSQL += strings.Join(dataSegment[:], ",")
    pSQL += " " + condition

    if nil != args {
        data = append(data, args...)
    }

    return s.exec(pSQL, data)
}

func (s *simpleMySQLDriver) Query(rowType interface{}, query string, args []interface{}) (rows []interface{}, err error) {
    sqlRows, err := s.db.Query(query, args...)
    if err != nil {
        return
    }
    defer func() {
       _ = sqlRows.Close()
    }()

    rows, err = simpleSQLDatabase.SQLRowsToStructureRows(sqlRows, rowType)
    return
}

func (s *simpleMySQLDriver) SimpleSelect(rowType interface{}, table string, col string, equalVal interface{}) (rows []interface{}, err error) {
    query := "select * from `" + table + "` where `" + col + "`=?"

    sqlRows, err := s.db.Query(query, equalVal)
    if err != nil {
        return
    }
    defer func() {
        sqlRows.Close()
    }()

    rows, err = simpleSQLDatabase.SQLRowsToStructureRows(sqlRows, rowType)

    return
}

func (s *simpleMySQLDriver) RowCount(table string, condition string, args []interface{}) (cnt uint64, err error) {
    pSQL := strings.Join([]string{
        "select count(*) from",
        "`" + table + "`",
        condition,
    }, " ")

    if args == nil {
        args = []interface{}{}
    }

    sqlRow, err := s.db.Query(pSQL, args...)
    if err != nil {
        return
    }
    defer func() {
        sqlRow.Close()
    }()

    for sqlRow.Next() {
        err = sqlRow.Scan(&cnt)
        break
    }

    return
}

func (s *simpleMySQLDriver) SimplePagingQuery(param simpleSQLDatabase.SimplePagingQueryParam) (ret *simpleSQLDatabase.SimplePagingQueryResult, err error) {

    count, err := s.RowCount(param.Table, param.Condition, param.ConditionArgs)
    if err != nil {
        return nil, err
    }

    start := param.Page * param.Count
    queryData := append(param.ConditionArgs, start, param.Count)

    pSQL := "select * from " +
        "`" + param.Table + "` " +
        param.Condition +
        " order by `c_id` desc limit ?,?"

    rows, err := s.Query(param.RowType, pSQL, queryData)
    if err != nil {
        return nil, err
    }

    result := &simpleSQLDatabase.SimplePagingQueryResult {
        Rows: rows,
        Total: count,
    }

    return result, err
}

func Load()  {
    enum.SimpleBuild(&Charsets)
}

func NewDriver(cfg interface{}) (driver simpleSQLDatabase.IDriver, err error) {
    s := &simpleMySQLDriver{}

    dbCfg, ok := cfg.(Config)
    if !ok {
        err = errors.New("incompatible config settings")
        return
    }

    openStr := dbCfg.User + ":" + dbCfg.Password + "@/" + dbCfg.DBName
    if dbCfg.Charset != "" {
        openStr += "?charset=" + dbCfg.Charset
    }

    s.db, err = sql.Open("mysql", openStr)

    if nil != err {
        return
    }

    if dbCfg.MaxConnection == 0 {
        dbCfg.MaxConnection = 100
    }

    s.db.SetMaxOpenConns(dbCfg.MaxConnection)
    s.db.SetMaxIdleConns(dbCfg.MaxConnection)

    driver = s
    return
}
