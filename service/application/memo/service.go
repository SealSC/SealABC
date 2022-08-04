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

package memo

import (
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/service/application/memo/memoInterface"
	"github.com/SealSC/SealABC/service/application/memo/memoSQLStorage"
	"github.com/SealSC/SealABC/service/application/memo/memoTables"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/storage/db"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

func Load() {
	memoInterface.Load()
	memoSQLStorage.Load()
	memoTables.Load()
}

func NewMemoApplication(config Config, tools crypto.Tools) (app chainStructure.IBlockchainExternalApplication, err error) {
	kvDriver, err := db.NewKVDatabaseDriver(config.KVDBName, config.KVDBConfig)
	if err != nil {
		log.Log.Error("can't load basic assets app for now: ", err.Error())
		return
	}

	var sqlDriver simpleSQLDatabase.IDriver = nil
	if config.EnableSQLDB {
		sqlDriver = config.SQLStorage
	}

	app = memoInterface.NewApplicationInterface(kvDriver, sqlDriver, tools)
	return
}
