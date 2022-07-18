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
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type levelDBDriver struct {
	db *leveldb.DB
}

type Config struct {
	DBFilePath string
	Options    opt.Options
}

func (l *levelDBDriver) Stat() (state interface{}, err error) {
	return l.db.GetProperty("leveldb.stats")
}

func (l *levelDBDriver) Close() {
	if nil != l.db {
		_ = l.db.Close()
	}
}

func NewDriver(cfg interface{}) (driver kvDatabase.IDriver, err error) {
	dbCfg, ok := cfg.(Config)

	if !ok {
		err = errors.New("incompatible config settings")
		return
	}

	db, err := leveldb.OpenFile(dbCfg.DBFilePath, &opt.Options{
		Filter: filter.NewBloomFilter(10),
	})

	if _, failed := err.(*errors.ErrCorrupted); failed {
		db, err = leveldb.RecoverFile(dbCfg.DBFilePath, nil)
	}

	if err != nil {
		return
	}

	l := &levelDBDriver{}
	l.db = db

	driver = l
	return
}
