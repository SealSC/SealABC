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

package memoInterface

import (
    "SealABC/crypto"
    "SealABC/dataStructure/enum"
    "SealABC/log"
    "SealABC/metadata/applicationResult"
    "SealABC/metadata/blockchainRequest"
    "SealABC/service"
    "SealABC/service/application/memo/memoSQLStorage"
    "SealABC/service/application/memo/memoSpace"
    "SealABC/service/system/blockchain/chainStructure"
    "SealABC/storage/db/dbInterface/kvDatabase"
    "SealABC/storage/db/dbInterface/simpleSQLDatabase"
    "encoding/json"
    "errors"
    "sync"
    "time"
)

type MemoApplication struct {
    operateLock sync.RWMutex
    poolLock    sync.Mutex

    reqPool  map[string] blockchainRequest.Entity


    CryptoTools crypto.Tools
    kvStorage   kvDatabase.IDriver
    sqlStorage  *memoSQLStorage.Storage
}

var applicationActions struct {
    Query   enum.Element
    Record  enum.Element
}

func (m *MemoApplication) Name() (name string) {
    return "Memo"
}

func (m *MemoApplication) kvQuery(req memoSpace.QueryRequest) (result interface{}, err error) {
    if len(req.Parameter) == 0 {
        err = errors.New("invalid parameter")
        return
    }
    result, err = m.QueryMemo(req.Parameter[0])
    return
}

func (m *MemoApplication) doRecord(req blockchainRequest.Entity) (result string, err error) {
    _, _, err = m.VerifyReq(req)
    if err != nil {
        log.Log.Error("verify memo failed: ", err.Error())
        return
    }

    m.poolLock.Lock()
    defer m.poolLock.Unlock()

    m.reqPool[req.Seal.HexHash()] = req
    return
}

func (m *MemoApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
    switch req.RequestAction {
    case applicationActions.Record.String():
        return m.doRecord(req)

    default:
        err = errors.New("service not support " + req.RequestAction + "@" + m.Name())
        return
    }
}

func (m *MemoApplication) Query(req string) (result interface{}, err error) {
    queryReq := memoSpace.QueryRequest{}
    err = json.Unmarshal([]byte(req), &queryReq)
    if err != nil {
        return
    }

    if m.sqlStorage != nil {
        return m.sqlStorage.DoQuery(queryReq)
    } else {
        return m.kvQuery(queryReq)
    }
}

func (m *MemoApplication) PreExecute(req blockchainRequest.Entity) (result []byte, err error) {
    _, _, err = m.VerifyReq(req)
    return
}

func (m *MemoApplication) Execute(
        req blockchainRequest.Entity,
        blockHeight uint64,
        actIndex uint32,
    ) (result applicationResult.Entity, err error) {

    _, memo, err := m.VerifyReq(req)
    if err != nil {
        log.Log.Error("verify memo failed: ", err.Error())
        return
    }

    memoJson, _ := json.Marshal(memo)

    m.operateLock.Lock()
    defer m.operateLock.Unlock()

    m.poolLock.Lock()
    defer func() {
        delete(m.reqPool, req.Seal.HexHash())
        m.poolLock.Unlock()
    }()

    err = m.kvStorage.Put(kvDatabase.KVItem{
        Key: memo.Seal.Hash,
        Data: memoJson,
    })

    if err != nil {
        log.Log.Error("save memo failed: ", err.Error())
        return
    }

    if m.sqlStorage != nil {
        _ = m.sqlStorage.StoreMemo(blockHeight, time.Now().Unix(), req, memo)
    }

    return
}

func (m *MemoApplication) Cancel(req blockchainRequest.Entity) (err error) {
    return
}

func (m *MemoApplication) RequestsForBlock() (reqList []blockchainRequest.Entity, cnt uint32) {
    m.operateLock.Lock()
    defer m.operateLock.Unlock()

    m.poolLock.Lock()
    defer m.poolLock.Unlock()

    for _, req := range m.reqPool {
        reqList = append(reqList, req)
    }

    cnt = uint32(len(reqList))
    return
}

func (m *MemoApplication) Information() (info service.BasicInformation) {
    info.Name = m.Name()
    info.Description = "this is a memo application"

    info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
    info.Api.Address = ""
    info.Api.ApiList = []service.ApiInterface {}
    return
}

func Load()  {
    enum.SimpleBuild(&applicationActions)
}

func NewApplicationInterface(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver , tools crypto.Tools) (app chainStructure.IBlockchainExternalApplication) {
    var storage *memoSQLStorage.Storage = nil
    if sqlDriver != nil {
        storage = memoSQLStorage.NewStorage(sqlDriver)
    }
    m := MemoApplication{
        CryptoTools: tools,
        kvStorage: kvDriver,
        sqlStorage: storage,
    }

    m.reqPool = map[string] blockchainRequest.Entity{}

    app = &m
    return
}
