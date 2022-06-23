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

package blockchain

import (
	"github.com/SealSC/SealABC/engine/engineService"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/service"
	"github.com/SealSC/SealABC/service/system/blockchain/chainApi"
	"github.com/SealSC/SealABC/service/system/blockchain/chainNetwork"
	"github.com/SealSC/SealABC/service/system/blockchain/chainSQLStorage"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/service/system/blockchain/serviceInterface"
)

func Load() {
	chainNetwork.Load()
	chainApi.Load()
	chainSQLStorage.Load()
}

func NewService(cfg Config) service.IService {
	//new blockchain
	chain := chainStructure.Blockchain{}

	//sql database
	var sqlStorage *chainSQLStorage.Storage = nil
	if cfg.SQLStorage != nil && cfg.EnableSQLDB {
		sqlStorage = &chainSQLStorage.Storage{
			Driver: cfg.SQLStorage,
		}
		chain.SetSQLStorage(sqlStorage)
	}

	err := chain.LoadBlockchain(cfg.Blockchain)
	if err != nil {
		log.Log.Error("create block chain failed")
		return nil
	}

	//new networks & service
	p2p := chainNetwork.NewNetwork(cfg.Network, &chain)

	//start api server
	apiServers := chainApi.NewServer(cfg.Api, &chain, p2p, sqlStorage)

	//register application executors
	for _, exe := range cfg.ExternalExecutors {
		_ = chain.Executor.RegisterApplicationExecutor(exe, &chain)
		apiServers.HttpJSON.Actions.RegisterApplicationQueryHandler(exe.Name(), exe.Query)
	}

	chainService := serviceInterface.NewServiceInterface(cfg.ServiceName, &chain, p2p, apiServers)

	//mount to engine
	_ = engineService.Mount(chainService)

	return chainService
}
