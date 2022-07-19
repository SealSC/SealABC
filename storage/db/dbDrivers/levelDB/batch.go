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

package levelDB

import (
	"errors"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (l *levelDBDriver) batchRead(kList [][]byte, needData bool) (kvList []kvDatabase.KVItem, err error) {
	hasErr := false
	for _, k := range kList {
		var v []byte
		var exists bool
		if needData {
			v, err = l.db.Get(k, nil)
			exists = err == nil
		} else {
			exists, err = l.db.Has(k, nil)
		}

		kv := kvDatabase.KVItem{
			Key:    k,
			Data:   v,
			Exists: exists,
		}

		kvList = append(kvList, kv)
		hasErr = hasErr && kv.Exists
	}

	return
}

func (l *levelDBDriver) BatchPut(kvList []kvDatabase.KVItem) (err error) {
	batch := new(leveldb.Batch)
	for _, kv := range kvList {
		batch.Put(kv.Key, kv.Data)
	}
	err = l.db.Write(batch, nil)
	return
}

func (l *levelDBDriver) BatchGet(kList [][]byte) (kvList []kvDatabase.KVItem, err error) {
	return l.batchRead(kList, true)
}

func (l *levelDBDriver) BatchDelete(kList [][]byte) (err error) {
	batch := new(leveldb.Batch)
	for _, k := range kList {
		batch.Delete(k)
	}
	err = l.db.Write(batch, nil)
	return
}

func (l *levelDBDriver) NewBatch() kvDatabase.Batch {
	return new(leveldb.Batch)
}

func (l *levelDBDriver) BatchWrite(b kvDatabase.Batch) (err error) {
	switch batch := b.(type) {
	case *leveldb.Batch:
		err = l.db.Write(batch, nil)
		return err
	default:
		err = errors.New("batch write error")
	}

	return
}

func (l *levelDBDriver) BatchCheck(kList [][]byte) (kvList []kvDatabase.KVItem, err error) {
	return l.batchRead(kList, false)
}

func (l *levelDBDriver) Traversal(condition []byte) (kvList []kvDatabase.KVItem) {
	iterator := l.db.NewIterator(util.BytesPrefix(condition), nil)
	defer iterator.Release()

	for iterator.Next() {
		newData := kvDatabase.KVItem{}

		k := iterator.Key()
		v := iterator.Value()
		newData.Key = make([]byte, len(k), len(k))
		newData.Data = make([]byte, len(v), len(v))
		copy(newData.Key, k)
		copy(newData.Data, v)
		newData.Exists = true

		kvList = append(kvList, newData)
	}

	return
}
