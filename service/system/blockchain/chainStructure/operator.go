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
    "SealABC/log"
    "SealABC/metadata/block"
    "SealABC/metadata/blockchainRequest"
    "SealABC/service/system/blockchain/chainTables"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "encoding/binary"
    "encoding/hex"
    "encoding/json"
    "errors"
)

const lastBlockKey = "lastBlockKey"

func (b *Blockchain) executeRequest(blk block.Entity) (err error) {
    for idx, req := range blk.Body.Requests {
        appRet, exeErr := b.Executor.ExecuteRequest(req, blk, uint32(idx))

        if exeErr != nil {
            err = exeErr
            break
        }

        app, _ := b.Executor.getExternalExecutor(req.RequestApplication)

        if b.SQLStorage != nil {
            var reqList []blockchainRequest.Entity
            var err error = nil
            if req.Packed {
                reqList, err = app.UnpackingActionsAsRequests(req)
                if err != nil {
                    log.Log.Errorf("unpack packed transaction for application %s failed: %s\r\n",
                        req.RequestApplication, err.Error())
                    continue
                }

                for _, r := range reqList {
                    sqlErr := b.SQLStorage.StoreTransaction(blk, r, appRet)
                    if sqlErr != nil {
                        log.Log.Error("store block in sql database failed: ", sqlErr.Error())
                    }

                    go b.SQLStorage.StoreAddress(blk, r)
                }
            } else {
                newReq := app.GetActionAsRequest(req)
                sqlErr := b.SQLStorage.StoreTransaction(blk, newReq, appRet)
                if sqlErr != nil {
                    log.Log.Error("store block in sql database failed: ", sqlErr.Error())
                }

                go b.SQLStorage.StoreAddress(blk, newReq)
            }
        }
        //todo: record result
    }

    return
}

func (b *Blockchain) AddBlock(blk block.Entity) (err error) {
    err = b.executeRequest(blk)
    if err != nil {
        log.Log.Error("execute requests in the block failed!")
        return
    }

    b.operateLock.Lock()
    defer b.operateLock.Unlock()

    blockBytes, err := json.Marshal(blk)
    if err != nil {
        return
    }

    heightKey := make([]byte, 8, 8)
    binary.BigEndian.PutUint64(heightKey, blk.Header.Height)

    err = b.Config.StorageDriver.BatchPut([]kvDatabase.KVItem {
        //first: key is height and data is block
        {
            Key: heightKey,
            Data: blockBytes,
        },

        //second: key is hash and data is height
        {
            Key: blk.Seal.Hash,
            Data: heightKey,
        },

        //last: save block as last block
        {
            Key: []byte(lastBlockKey),
            Data: blockBytes,
        },
    })

    if err != nil {
        return
    }

    b.currentHeight = blk.Header.Height
    b.lastBlock = &blk

    if b.SQLStorage != nil {
        go func() {
            _ = b.SQLStorage.StoreBlock(blk)
        }()
    }

    return
}

func (b *Blockchain) GetBlockByHeight(height uint64) (blk block.Entity, err error) {
    b.operateLock.RLock()
    defer b.operateLock.RUnlock()

    heightKey := make([]byte, 8, 8)
    binary.BigEndian.PutUint64(heightKey, height)
    kv, err := b.Config.StorageDriver.Get(heightKey)
    if err != nil {
        return
    }

    if !kv.Exists {
        err = errors.New("no such block")
        return
    }

    err = json.Unmarshal(kv.Data, &blk)

    return
}

func (b *Blockchain) getBlockFromKVDBByHeight(height uint64) (blk chainTables.BlockListRow, err error)  {
    blkEntity, err := b.GetBlockByHeight(height)
    if err != nil {
        return
    }
    blk.FromBlockEntity(blkEntity)
    return
}

func (b *Blockchain) GetBlockRowByHeight(height uint64) (blk chainTables.BlockListRow, err error) {
    if b.SQLStorage != nil {
        return b.SQLStorage.GetBlock(height)
    } else {
        return b.getBlockFromKVDBByHeight(height)
    }
}

func (b *Blockchain) getBlockFromKVDBByHash(hash []byte) (blk chainTables.BlockListRow, err error) {
    b.operateLock.RLock()
    defer b.operateLock.RUnlock()

    //get height by hash first
    kvHeight, err := b.Config.StorageDriver.Get(hash)
    if err != nil {
        return
    }

    if !kvHeight.Exists {
        err = errors.New("no such block")
        return
    }

    //get block by height finally
    kvBlock, err := b.Config.StorageDriver.Get(kvHeight.Data)
    if err != nil {
        return
    }

    if !kvBlock.Exists {
        err = errors.New("no such block")
        return
    }

    blkEntity := block.Entity{}
    err = json.Unmarshal(kvBlock.Data, &blkEntity)
    if err != nil {
        return
    }

    blk.FromBlockEntity(blkEntity)
    return
}

func (b *Blockchain) GetBlockRowByHash (hash string) (blk chainTables.BlockListRow, err error) {
    hashBytes, err := hex.DecodeString(hash)
    if err != nil {
        return
    }
    if b.SQLStorage != nil {
        return b.SQLStorage.GetBlockByHash(hash)
    } else {
        return b.getBlockFromKVDBByHash(hashBytes)
    }
}

func (b *Blockchain) GetLastBlock() (last *block.Entity) {
    kv, _ := b.Config.StorageDriver.Get([]byte(lastBlockKey))

    if !kv.Exists {
        return
    }

    lastBlk := block.Entity{}
    err := json.Unmarshal(kv.Data, &lastBlk)
    if err != nil {
        return
    }

    last = &lastBlk
    return
}

func (b *Blockchain) CurrentHeight() uint64 {
    return b.currentHeight
}
