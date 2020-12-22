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

package chainSQLStorage

import (
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/applicationResult"
    "github.com/SealSC/SealABC/metadata/block"
    "github.com/SealSC/SealABC/metadata/blockchainRequest"
    "github.com/SealSC/SealABC/service/system/blockchain/chainTables"
    "encoding/hex"
)

func (s *Storage) StoreBlock(blk block.Entity) (err error)  {
    rows := chainTables.BlockList.NewRows().(chainTables.BlockListRows)
    rows.InsertBlock(blk)

    _, err = s.Driver.Insert(&rows, true)
    if err != nil {
        log.Log.Error("insert to sql database failed: ", err.Error())
    }
    return
}

func (s *Storage) StoreTransaction(blk block.Entity, req blockchainRequest.Entity, result applicationResult.Entity) (err error) {
    rows := chainTables.Requests.NewRows().(chainTables.RequestRows)
    rows.InsertTransaction(blk, req, result)

    _, err = s.Driver.Insert(&rows, true)
    if err != nil {
        log.Log.Error("insert transaction to block transaction list failed: ", err.Error())
    }

    return
}

func (s *Storage) StoreAddress(blk block.Entity, req blockchainRequest.Entity) (err error) {
    defer func() {
        if r := recover(); r != nil {
            log.Log.Error("got panic: ", r)
        }

        if err != nil {
            log.Log.Error(err.Error())
            return
        }
    }()

    cnt, err := s.Driver.RowCount(
        chainTables.AddressList.Name(),
        "where `c_address`=?",
        []interface{}{
            hex.EncodeToString(req.Seal.SignerPublicKey),
        })


    rows := chainTables.AddressList.NewRows().(chainTables.AddressListRows)
    rows.UpdateAddress(blk, req)
    if cnt == 0 {
        _, err = s.Driver.Insert(&rows, true)
    } else {
        _, err = s.Driver.Replace(&rows)
    }

    return
}
