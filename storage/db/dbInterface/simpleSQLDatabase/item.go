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
    "reflect"
    "strconv"
)

type ITable interface {
    Name() string
    Columns() []string
    ColumnsForInsert() []string
    FieldForInsert() []string
    NewRows() interface{}
}

type BasicTable struct {
    columns             []string
    columnsForInsert    []string
    insertField         []string

    Instance            interface{}
}

func (b BasicTable) Name() string {
    return ""
}

func (b BasicTable) Columns() (list []string) {
    if b.Instance == nil {
        return
    }

    if len(b.columns) == 0 {
        list, _ = ColumnsFromTag(b.Instance, false)
        b.columns = list
    }

    return b.columns
}

func (b BasicTable) ColumnsForInsert() (list []string) {
    if b.Instance == nil {
        return
    }

    if len(b.columnsForInsert) == 0 {
        list, _ = ColumnsFromTag(b.Instance, true)
        b.columnsForInsert = list
    }

    return b.columnsForInsert
}

func (b BasicTable) FieldForInsert() (fields []string){
    if b.Instance == nil {
        return
    }

    if len(b.insertField) == 0 {
        fields, _ = FieldsFromLocalTable(b.Instance, true)
        b.insertField = fields
    }

    return b.insertField
}

type IRows interface {
    Table() ITable
    Count() int
    DataForInsert() []interface{}
    Data(requiredColumns []string) []interface{}
}

type BasicRows struct {
    Rows []interface{}

    Instance IRows
}

func (b BasicRows) Count() int {
    return len(b.Rows)
}

func (b BasicRows) Data(requiredField []string) (data []interface{}) {
    if requiredField == nil {
        return
    }

    for _, d := range b.Rows {
        list := StringDataFromRow(d, requiredField)
        data = append(data, list...)
    }

    return
}

func (b BasicRows) DataForInsert() (data []interface{}) {
    if b.Instance == nil {
        return
    }

    table := b.Instance.Table()
    if table == nil {
        return
    }

    fieldForInsert := table.FieldForInsert()
    for _, d := range b.Rows {
        list := StringDataFromRow(d, fieldForInsert)
        data = append(data, list...)
    }

    return
}

type SimplePagingQueryParam struct {
    Page          int64
    Count         int64
    Table         string
    RowType       interface{}
    Condition     string
    ConditionArgs []interface{}
}

func (s *SimplePagingQueryParam) PageFromString(page string)  {
    s.Page = 0
    if page == "" {
        return
    }

    p, err := strconv.ParseInt(page, 10, 64)
    if err != nil {
        return
    }

    s.Page = p
}

type SimplePagingQueryResult struct {
    Total uint64
    Rows  interface{}
}

func NewRowsInstance(base interface{}) (rows interface{}) {
    rowsType := reflect.TypeOf(base)
    newRows := reflect.New(rowsType).Elem()

    if !newRows.CanAddr() {
        return
    }

    field := newRows.FieldByName("Instance")
    if !field.IsValid() {
        return
    }

    if !field.CanSet() {
        return
    }

    field.Set(newRows.Addr())

    return newRows.Interface()
}
