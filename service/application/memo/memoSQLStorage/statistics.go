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

type  countStatistics struct {
    Count   uint64
}

type distributed struct {
    Count   uint64
    Group   string
}

type statistic struct {
    RecorderDistributed []interface{}
    TypeDistributed     []interface{}
    SizeDistributed     []interface{}
    TotalMemo           uint64
    TotalRecorder       uint64
}

func (s *Storage) GetStatistics(_ []string) (ret interface{}, err error) {
    memoCount, err := s.Driver.RowCount(`t_memo_list`, "", nil)
    if err != nil {
        return
    }

    addressCount, err := s.Driver.RowCount(`t_memo_address_list`, "", nil)
    if err != nil {
        return
    }

    pSQL := "SELECT COUNT(*) as `count`, `c_recorder` FROM `t_memo_list` WHERE 1 GROUP BY `c_recorder` ORDER BY `count` DESC limit 0, 20"
    recorderRows, err := s.Driver.Query(distributed{}, pSQL, nil)
    if err != nil {
        return
    }

    pSQL = "SELECT COUNT(*) as `count`, `c_type`  FROM `t_memo_list` WHERE 1 GROUP BY `c_type` ORDER BY `count` DESC limit 0, 20"
    typeRows, err := s.Driver.Query(distributed{}, pSQL, nil)
    if err != nil {
        return
    }

    pSQL = "SELECT COUNT(*) as `count`, `c_size` FROM `t_memo_list` WHERE 1 GROUP BY `c_size` ORDER BY `count` DESC limit 0, 20"
    sizeRows, err := s.Driver.Query(distributed{}, pSQL, nil)
    if err != nil {
        return
    }

    ret = statistic {
        RecorderDistributed: recorderRows,
        TypeDistributed: typeRows,
        SizeDistributed: sizeRows,
        TotalMemo: memoCount,
        TotalRecorder: addressCount,
    }

    return
}
