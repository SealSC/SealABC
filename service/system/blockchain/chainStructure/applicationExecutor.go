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
    "SealABC/metadata/applicationResult"
    "SealABC/metadata/block"
    "SealABC/metadata/blockchainRequest"
    "SealABC/service"
    "errors"
    "sync"
)

type applicationExecutor struct {
    ExternalExecutors   map[string]IBlockchainExternalApplication

    externalExeLock     sync.RWMutex
}

type IBlockchainExternalApplication interface {

    Name() (name string)

    //receive a request from user
    PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error)

    //query
    Query(req []byte) (result interface{}, err error)

    //receive and verify request from a block
    PreExecute(req blockchainRequest.Entity, header block.Entity) (result []byte, err error)

    //receive and execute confirmed block
    Execute(req blockchainRequest.Entity, header block.Entity, actIndex uint32) (result applicationResult.Entity, err error)

    //handle consensus failed
    Cancel(req blockchainRequest.Entity) (err error)

    //build request list for new block
    RequestsForBlock(block block.Entity) (entity []blockchainRequest.Entity, cnt uint32)

    //external service information
    Information() (info service.BasicInformation)

    SetBlockchainService(bs interface{})
}

func (a *applicationExecutor)getExternalExecutor(name string) (exe IBlockchainExternalApplication, err error) {
    exe, exists := a.ExternalExecutors[name]
    if !exists {
        err = errors.New("no applicationExecutor named " + name)
        return
    }

    return
}

func (a *applicationExecutor) RegisterApplicationExecutor(s IBlockchainExternalApplication, bs interface{}) (err error)  {
    a.externalExeLock.Lock()
    defer a.externalExeLock.Unlock()

    if s == nil {
        return
    }

    if exe, _ := a.getExternalExecutor(s.Name()); exe != nil {
        return
    }

    s.SetBlockchainService(bs)
    a.ExternalExecutors[s.Name()] = s
    return
}

func (a *applicationExecutor)PreExecute(act blockchainRequest.Entity, blk block.Entity) (result []byte, err error) {
    a.externalExeLock.RLock()
    defer a.externalExeLock.RUnlock()

    exe, err := a.getExternalExecutor(act.RequestApplication)
    if err != nil {
        return
    }

    return exe.PreExecute(act, blk)
}

func (a *applicationExecutor)ExecuteRequest(req blockchainRequest.Entity, blk block.Entity, actIndex uint32) (result applicationResult.Entity, err error) {
    a.externalExeLock.RLock()
    defer a.externalExeLock.RUnlock()

    exe, err := a.getExternalExecutor(req.RequestApplication)
    if err != nil {
        return
    }

    return exe.Execute(req, blk, actIndex)
}

func (a *applicationExecutor)GetRequestListToBuildBlock(block block.Entity) (reqList []blockchainRequest.Entity) {
    a.externalExeLock.RLock()
    defer a.externalExeLock.RUnlock()

    for _, exe := range a.ExternalExecutors {
        req, cnt := exe.RequestsForBlock(block)
        if cnt == 0 {
            continue
        }

        reqList = append(reqList, req...)
    }

    return
}

func (a *applicationExecutor)PushRequest(req blockchainRequest.Entity) (result interface{}, err error) {
    a.externalExeLock.RLock()
    defer a.externalExeLock.RUnlock()

    exe, err := a.getExternalExecutor(req.RequestApplication)
    if err != nil {
        return
    }


    return exe.PushClientRequest(req)
}

func (a *applicationExecutor)ExternalApplicationInformation() (list []service.BasicInformation) {
    a.externalExeLock.RLock()
    defer a.externalExeLock.RUnlock()

    for _, exe := range a.ExternalExecutors {
        info := exe.Information()
        list = append(list, info)
    }

    return
}
