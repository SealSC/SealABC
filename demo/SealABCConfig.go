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

package main

import (
    "github.com/SealSC/SealABC/cli"
    "github.com/SealSC/SealABC/network/http"
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
)

type Config struct {
    ConsensusServiceAddress string   `json:"consensus_service_address"`
    ConsensusMember         []string `json:"consensus_member"`

    ConsensusDisabled       bool `json:"consensus_disabled"`

    BlockchainServiceAddress string `json:"blockchain_service_address"`
    BlockchainApiConfig      http.Config `json:"blockchain_api_config"`
    BlockchainServiceSeeds   []string `json:"blockchain_service_seeds"`

    SelfKey     string `json:"self_key"`
    Members     []string `json:"members"`

    ChainDB       string `json:"chain_db"`
    LedgerDB      string `json:"ledger_db"`
    MemoDB        string `json:"memo_db"`
    SmartAssetsDB string `json:"smart_assets_db"`
    TraceableStorageDB string `json:"traceable_storage_db"`

    LogFile     string `json:"log_file"`
    LogLevel    uint32 `json:"log_level"`

    EngineApiConfig     http.Config `json:"engine_api_config"`

    EnableSQLStorage bool `json:"enable_sql_storage"`

    MySQLDBName string `json:"mysql_db_name"`
    MySQLUser   string `json:"mysql_user"`
    MySQLPwd    string  `json:"mysql_pwd"`
}

func (c *Config) Load() (err error) {
    cfgStr, err := ioutil.ReadFile(cli.Parameters.ConfigFile)

    if nil != err {
        fmt.Println("load config file error: ", err)
        return
    }

    cfgStr = bytes.TrimPrefix(cfgStr, []byte("\xef\xbb\xbf"))
    err = json.Unmarshal(cfgStr, c)

    return
}

var SealABCConfig = Config {
    LogFile:            "github.com/SealSC/SealABC.log",
    LogLevel:           5,
}
