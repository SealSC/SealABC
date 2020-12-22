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

package serviceInterface

import (
    "github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/block"
    "github.com/SealSC/SealABC/metadata/serviceRequest"
    "github.com/SealSC/SealABC/service"
    "github.com/SealSC/SealABC/service/system/blockchain/chainApi"
    "github.com/SealSC/SealABC/service/system/blockchain/chainNetwork"
    "github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
    "bytes"
    "encoding/json"
    "errors"
    "sync"
)

const defaultServiceName = "SEAL-BLOCKCHAIN-SYSTEM"

type BlockchainService struct {
    serviceName string
    syncLock    sync.Mutex

    chain       *chainStructure.Blockchain
    p2pService  *chainNetwork.P2PService

    apiServers  *chainApi.ApiServers
}

func (b BlockchainService) getNewBlockRequestFromConsensus(data interface{}) (blk block.Entity, err error) {
    srvData, ok := data.(serviceRequest.Entity)
    if !ok {
        err = errors.New("invalid request data")
        log.Log.Error("invalid request data")
        return
    }

    err = json.Unmarshal([]byte(srvData.Payload), &blk)
    if err != nil {
        log.Log.Error("unmarshal block json failed: ", err.Error())
        return
    }

    blockBytes, err := structSerializer.ToMFBytes(blk)
    if err != nil {
        log.Log.Error("serialize block to verify signature failed: ", err.Error())
        return
    }

    _, err = srvData.CustomerSeal.Verify(blockBytes, b.chain.Config.CryptoTools.HashCalculator)
    if err != nil {
        log.Log.Error("invalid request data signature: ", err.Error())
        return
    }

    return
}

//service name
func (b *BlockchainService) Name() (name string) {
    return b.serviceName
}

//build a new block for consensus
func (b *BlockchainService) RequestsForConsensus() (req [][]byte, cnt uint32) {
    if chainNetwork.Syncing {
        return
    }

    newBlankBlock := b.chain.NewBlankBlock()
    reqList := b.chain.Executor.GetRequestListToBuildBlock(newBlankBlock)
    newBlock := b.chain.NewBlock(reqList, newBlankBlock)

    blockBytes, err := structSerializer.ToMFBytes(newBlock)
    if err != nil {
        panic("blockchain data structure damaged when serialize to sign")
    }

    blockJson, err := json.Marshal(newBlock)
    if err != nil {
        panic("blockchain data structure damaged when serialize to json")
    }

    newBlockReq := serviceRequest.Entity{}
    cnt = 1

    newBlockReq.RequestService = b.Name()
    newBlockReq.Payload = string(blockJson)
    _ = newBlockReq.CustomerSeal.Sign(blockBytes, b.chain.Config.CryptoTools, b.chain.Config.Signer.PrivateKeyBytes())

    reqBytes, err := json.Marshal(newBlockReq)
    if err != nil {
        return
    }
    req = append(req, reqBytes)

    //log.Log.Println("new block @", newBlock.Header.Height, " for consensus, data length: ", len(reqBytes))
    return
}

func (b *BlockchainService) verifyBlock(blk block.Entity) (err error) {
    //verify block
    _, err = blk.Verify(b.chain.Config.CryptoTools)
    if err != nil {
        log.Log.Error("block @", blk.Header.Height, " verify failed")
        return
    }

    //genesis block has no prev-block
    if blk.Header.Height == 0 {
        return
    }

    lastBlk := b.chain.GetLastBlock()
    if blk.Header.Height != lastBlk.Header.Height + 1 {
        log.Log.Warn("unable to verify the block: local height@", b.chain.CurrentHeight(), " new block@", blk.Header.Height )
        err = errors.New("unable to verify the block because the current block cannot be aligned to the current height")
        return
    }

    if !bytes.Equal(blk.Header.PrevBlock, lastBlk.Seal.Hash) {
        log.Log.Warn("new block's prev-block is not on chain last block: ",blk.Seal.Hash, " new block's prev: ", blk.Header.PrevBlock )
        err = errors.New("new block's prev-block is not on chain")
        return
    }

    return
}

//handle the new block received from consensus
func (b *BlockchainService) PreExecute(data interface{}) (result []byte, err error) {
    blk, err := b.getNewBlockRequestFromConsensus(data)
    if err != nil {
        return
    }

    err = b.verifyBlock(blk)
    if err != nil {
        return
    }

    //pre-executeRequest the request
    for _, req := range blk.Body.Requests {
        result, err = b.chain.Executor.PreExecute(req, blk)
        //todo: handle the result
    }

    return
}

//storage a new block
func (b *BlockchainService) Execute(data interface{}) (result []byte, err error){
    blk, err := b.getNewBlockRequestFromConsensus(data)
    if err != nil {
        return
    }

    if blk.Header.Height > b.chain.CurrentHeight() + 1 {
        nodes := b.p2pService.NetworkService.GetAllLinkedNode()
        if len(nodes) == 0 {
            log.Log.Error("no blockchain service neighbors!")
            return
        }

        log.Log.Warn("start sync block! @ service: ", b.Name())
        go b.p2pService.StartSync(nodes, blk.Header.Height)
        return
    }

    err = b.verifyBlock(blk)
    if err != nil {
        return
    }

    //execute request
    //for idx, req := range blk.Body.Requests {
    //    appRet, exeErr := b.chain.Executor.ExecuteRequest(req, blk.Header.Height, uint32(idx))
    //
    //    if exeErr != nil {
    //        err = exeErr
    //        break
    //    }
    //
    //    if b.chain.SQLStorage != nil {
    //        _ = b.chain.SQLStorage.StoreTransaction(blk, req, appRet)
    //    }
    //    //todo: record result
    //}
    //
    //if err != nil {
    //    return
    //}

    err = b.chain.AddBlock(blk)
    if err != nil {
        log.Log.Error("add block @", blk.Header.Height, " failed.")
    } else {
        log.Log.Println("block @", blk.Header.Height, " complete.")
    }
    return
}

//placeholder for action rollback
func (b *BlockchainService) Cancel(srvData interface{}) (err error) {
    //todo: roll back the action if needed (maybe re-sync the block from a right height is the easiest way)
    return
}

func (b *BlockchainService) Information() (info service.BasicInformation) {
    info.Name = b.Name()
    info.Description = "this is a blockchain service."
    info.Api.Protocol = service.ApiProtocols.HTTP.String()
    info.Api.Address = b.apiServers.HttpJSON.Address()
    info.Api.ApiList = b.apiServers.HttpJSON.Actions.Information()
    info.SubServices = b.chain.Executor.ExternalApplicationInformation()

    return
}

func NewServiceInterface(
        name string,
        chain *chainStructure.Blockchain,
        p2p *chainNetwork.P2PService,
        apiServers *chainApi.ApiServers,
    ) service.IService {

    if name == "" {
        name = defaultServiceName
    }

    bs :=  &BlockchainService {
        serviceName:    name,
        chain:          chain,
        p2pService:     p2p,
        apiServers:     apiServers,
    }

    return bs
}
