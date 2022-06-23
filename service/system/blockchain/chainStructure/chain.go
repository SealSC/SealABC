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

package chainStructure

import (
	"github.com/SealSC/SealABC/dataStructure/state"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/service/system/blockchain/chainSQLStorage"
	"sync"
)

type Blockchain struct {
	Config   Config
	Executor applicationExecutor

	lastBlock     *block.Entity
	SQLStorage    *chainSQLStorage.Storage
	currentHeight uint64
	operateLock   sync.RWMutex

	StateCache state.Database
}

func (b *Blockchain) SetSQLStorage(sqlStorage *chainSQLStorage.Storage) {
	b.SQLStorage = sqlStorage
}

func (b *Blockchain) LoadBlockchain(cfg Config) (err error) {
	b.Config = cfg
	b.Executor.ExternalExecutors = map[string]IBlockchainExternalApplication{}

	lastBlock := b.GetLastBlock()
	if lastBlock == nil {
		b.currentHeight = 0
		return
	}

	log.Log.Println("get latest block: ", lastBlock.Header.Height)

	b.lastBlock = lastBlock
	b.currentHeight = lastBlock.Header.Height
	return
}
