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

import (
    "database/sql"
    "reflect"
)

const (
    ColumnNameTag = "col"
    DefaultValueTag = "def"
    IgnoreForInsertTag = "ignoreInsert"
)

func ColumnsFromTag(table interface{}, forInsert bool) (columns [] string, err error){
    tType := reflect.TypeOf(table)

    if reflect.Struct != tType.Kind() {
        return
    }

    for i := 0; i < tType.NumField(); i++ {
        t := tType.Field(i)

        if forInsert {
            ignoreForInsert := t.Tag.Get(IgnoreForInsertTag)
            if ignoreForInsert == "true" {
                continue
            }
        }

        col := t.Tag.Get(ColumnNameTag)
        if "" == col {
            continue
        }

        columns = append(columns, col)
    }

    return
}

func FieldsFromLocalTable(table interface{}, forInsert bool) (fields [] string, err error){
    tType := reflect.TypeOf(table)

    if reflect.Struct != tType.Kind() {
        return
    }

    for i := 0; i < tType.NumField(); i++ {
        t := tType.Field(i)

        if forInsert {
            ignoreForInsert := t.Tag.Get(IgnoreForInsertTag)
            if ignoreForInsert == "true" {
                continue
            }
        }

        f := t.Name
        fields = append(fields, f)
    }

    return
}

func StringDataFromRow(row interface{}, requiredField []string) (data []interface{}) {
    rType := reflect.TypeOf(row)

    if reflect.Struct != rType.Kind() {
        return
    }

    rValue := reflect.ValueOf(row)

    for _, f := range requiredField {
        sf, exist := rType.FieldByName(f)
        if !exist {
            continue
        }
        v := rValue.FieldByName(f)
        if reflect.String != v.Kind() {
            continue
        }

        val := v.String()
        if "" == val {
            defaultVal := sf.Tag.Get(DefaultValueTag)
            val = defaultVal
        }
        data = append(data, val)
    }

    return
}

func isValidColumn(colInRow []string, field reflect.StructField) bool  {
    valid := false
    for _, col := range colInRow {
        colName := field.Tag.Get(ColumnNameTag)
        valid = (col == colName) || ("" == colName)
        if valid {
            break
        }
    }

    return valid
}
func SQLRowsToStructureRows(sqlRows *sql.Rows, targetRowType interface{}) (rows []interface{}, err error) {
    tType := reflect.TypeOf(targetRowType)
    if reflect.Struct != tType.Kind() {
        return
    }

    colInRows, err := sqlRows.Columns()
    if nil != err {
        return
    }

    for sqlRows.Next() {
        newItem := reflect.New(tType).Elem()

        var valList []interface{}

        for i :=0; i < newItem.NumField(); i++ {
            colVal := newItem.Field(i)
            if !colVal.CanAddr() {
                continue
            }

            if !isValidColumn(colInRows, tType.Field(i)) {
                continue
            }

            valList = append(valList, colVal.Addr().Interface())
        }

        err = sqlRows.Scan(valList...)
        if nil != err {
            break
        }
        rows = append(rows, newItem.Interface())
    }

    return
}


