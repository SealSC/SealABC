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

package oracleInterface

import (
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"encoding/json"
)

const keyPrefix = "oracle_map_"

type SortMap struct {
	kvDB kvDatabase.IDriver
}

func NewSortMap(kvDB kvDatabase.IDriver) *SortMap {
	return &SortMap{kvDB: kvDB}
}

func (s *SortMap) set(k string, v blockchainRequest.Entity) error {
	marshal, err := json.Marshal(v)
	if err != nil {
		return err
	}
	err = s.kvDB.Put(kvDatabase.KVItem{
		Key:  []byte(keyPrefix + k),
		Data: marshal,
	})
	return err
}

func (s *SortMap) get(k string) (result blockchainRequest.Entity, ok bool, err error) {
	var kv kvDatabase.KVItem
	kv, err = s.kvDB.Get([]byte(keyPrefix + k))
	if err != nil {
		return
	}
	ok = kv.Exists
	if !ok {
		return
	}
	err = json.Unmarshal(kv.Data, &result)
	if err != nil {
		return
	}
	return
}

func (s *SortMap) list() (result []blockchainRequest.Entity) {
	bytes := []byte(keyPrefix)
	traversal := s.kvDB.Traversal(bytes)
	result = make([]blockchainRequest.Entity, len(traversal))
	for i, item := range traversal {
		if !item.Exists {
			break
		}
		_ = json.Unmarshal(item.Data, &result[i])
	}
	return result
}

func (s *SortMap) del(k string) error {

	return s.kvDB.Delete([]byte(keyPrefix + k))
}
