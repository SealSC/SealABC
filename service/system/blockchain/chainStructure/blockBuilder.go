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
	"github.com/SealSC/SealABC/dataStructure/merkleTree"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"time"
)

func (b *Blockchain) buildBasicBlock(requests []blockchainRequest.Entity) (newBlock block.Entity) {
	newBlock = block.Entity{}

	newBlock.Header.Version = "1"
	newBlock.Body.RequestsCount = len(requests)

	for _, req := range requests {
		if req.Packed {
			newBlock.Body.RequestsCount += int(req.PackedCount - 1)
		}
	}

	newBlock.Body.Requests = requests
	newBlock.Header.Timestamp = uint64(time.Now().Unix())

	mt := merkleTree.Tree{}
	for _, req := range requests {
		mt.AddHash(req.Seal.Hash)
	}

	rt, err := mt.Calculate()
	if err != nil {
		log.Log.Error("calc merkle  tree failed: ", err.Error())
	}
	newBlock.Header.TransactionsRoot = rt

	return
}

func (b *Blockchain) NewBlankBlock() (newBlock block.Entity) {
	newBlock = b.buildBasicBlock(nil)

	if b.lastBlock != nil {
		newBlock.Header.Height = b.lastBlock.Header.Height + 1

		//set block prev hash
		newBlock.Header.PrevBlock = append([]byte{}, b.lastBlock.Seal.Hash...)
		newBlock.Header.StateRoot = b.lastBlock.Header.StateRoot
	}

	//set block hash
	err := newBlock.Sign(b.Config.CryptoTools, b.Config.Signer.PrivateKeyBytes())
	if err != nil {
		log.Log.Error("sign blank block failed: ", err.Error())
	}

	newBlock.BlankSeal = newBlock.Seal
	return
}

func (b *Blockchain) NewBlock(requests []blockchainRequest.Entity, blankBlock block.Entity) (newBlock block.Entity) {
	//build a basic block
	newBlock = b.buildBasicBlock(requests)
	newBlock.Header.Timestamp = blankBlock.Header.Timestamp
	newBlock.BlankSeal = blankBlock.BlankSeal
	newBlock.Header.StateRoot = blankBlock.Header.StateRoot
	//set block height
	if b.lastBlock != nil {
		newBlock.Header.Height = b.lastBlock.Header.Height + 1

		//set block prev hash
		newBlock.Header.PrevBlock = append([]byte{}, b.lastBlock.Seal.Hash...)
	}

	//set block hash
	err := newBlock.Sign(b.Config.CryptoTools, b.Config.Signer.PrivateKeyBytes())
	if err != nil {
		log.Log.Error("sign new block failed: ", err.Error())
	}

	return
}
