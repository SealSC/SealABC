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

package db

import (
    "SealABC/storage/db/dbDrivers/levelDB"
    "SealABC/storage/db/dbDrivers/simpleMysql"
    "SealABC/storage/db/dbInterface"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

type kvDriverLoader func(cfg interface{}) (engine kvDatabase.IDriver, err error)
type simpleSQLDriverLoader func(cfg interface{}) (engine simpleSQLDatabase.IDriver, err error)
var kvDBDriverLoader = map[string]kvDriverLoader{
    dbInterface.LevelDB: levelDB.NewDriver,
}

var simpleSQLDBDriverLoader = map[string] simpleSQLDriverLoader {
    dbInterface.MySQL: simpleMysql.NewDriver,
}

func Load()  {
    simpleMysql.Load()
}

func NewKVDatabaseDriver(name string, cfg interface{}) (driver kvDatabase.IDriver, err error){
    driverLoader, supported := kvDBDriverLoader[name]
    if !supported {
        return
    }

    driver, err = driverLoader(cfg)
    return
}

func NewSimpleSQLDatabaseDriver(name string, cfg interface{}) (driver simpleSQLDatabase.IDriver, err error)  {
    driverLoader, supported := simpleSQLDBDriverLoader[name]
    if !supported {
        return
    }

    driver, err = driverLoader(cfg)
    return
}
