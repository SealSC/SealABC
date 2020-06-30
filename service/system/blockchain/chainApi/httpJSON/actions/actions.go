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

package actions

import (
    "SealABC/network/http"
    "SealABC/service"
    "SealABC/service/system/blockchain/chainSQLStorage"
    "SealABC/service/system/blockchain/chainStructure"
    "github.com/gin-gonic/gin"
    "SealABC/log"
    "SealABC/dataStructure/enum"
    "SealABC/service/system/blockchain/chainNetwork"
)

var URLParameterKeys = struct {
    HexHash  enum.Element
    Height   enum.Element

    Page    enum.Element
    Count   enum.Element

    App     enum.Element
    Action  enum.Element

    Type    enum.Element
    Param   enum.Element
}{}

type apiHandler interface {
    http.IRequestHandler
    buildUrlPath() string
    setServerInfo(basePath string,
        chain *chainStructure.Blockchain,
        p2p *chainNetwork.P2PService,
        sqlStorage *chainSQLStorage.Storage,
        queryHandler map[string] applicationQueryHandler)
}

type baseHandler struct {
    serverBasePath  string
    chain           *chainStructure.Blockchain
    p2p             *chainNetwork.P2PService
    sqlStorage      *chainSQLStorage.Storage
    appQueryHandler map[string] applicationQueryHandler
}

func (b *baseHandler) setServerInfo(basePath string,
    chain *chainStructure.Blockchain,
    p2p *chainNetwork.P2PService,
    sqlStorage *chainSQLStorage.Storage,
    queryHandler map[string] applicationQueryHandler) {

    b.serverBasePath = basePath
    b.chain = chain
    b.p2p = p2p
    b.sqlStorage = sqlStorage
    b.appQueryHandler = queryHandler
}

type applicationQueryHandler func([]byte) (interface{}, error)
type ChainApiActions struct {
    serverBase          string
    actionList          []apiHandler
    appQueryHandler     map[string] applicationQueryHandler

    chain           *chainStructure.Blockchain
    p2p             *chainNetwork.P2PService

    apiInformation  []service.ApiInterface
    sqlStorage      *chainSQLStorage.Storage
}

func (c *ChainApiActions) RouteRegister(router gin.IRouter) {
    c.serverBase = "/api/v1"
    gr := router.Group(c.serverBase)
    {
        for _, h := range c.actionList {
            h.setServerInfo(c.serverBase, c.chain, c.p2p, c.sqlStorage, c.appQueryHandler)
            h.RouteRegister(gr)
        }
    }
}

func (c *ChainApiActions) BasicInformation() (info http.HandlerBasicInformation) {
    log.Log.Println("this is placeholder for BasicInformation method.")
    return
}

func (c *ChainApiActions) Handle(_ *gin.Context) {
    log.Log.Println("this is placeholder for Handle method.")
    return
}

func NewActions(basePath string, chain *chainStructure.Blockchain, p2p *chainNetwork.P2PService, sqlStorage *chainSQLStorage.Storage) *ChainApiActions {
    action := ChainApiActions{}

    action.serverBase = basePath
    action.chain = chain
    action.p2p = p2p
    action.sqlStorage = sqlStorage

    action.actionList = []apiHandler {
        &callApplication{},
        &getBlockByHash{},
        &getBlockByHeight{},
        &getTransactions{},
        &queryApplication{},
    }
    action.appQueryHandler = map[string] applicationQueryHandler {}

    if sqlStorage != nil {
        action.actionList = append(action.actionList,
            &getBlockList{},
            &getTransactionByHash{},
            &getTransactionByApplicationAndAction{},
            &getAddressList{},
            &getTransactionByHeight{})
    }

    return &action
}

func (c *ChainApiActions) RegisterApplicationQueryHandler(exeName string, handler applicationQueryHandler)  {
    c.appQueryHandler[exeName] = handler
}

func (c *ChainApiActions)Information() []service.ApiInterface {
    if len(c.apiInformation) != 0 {
        return c.apiInformation
    }

    for _, act := range c.actionList {
        actInfo := act.BasicInformation()
        ai := service.ApiInterface{}

        ai.Description = actInfo.Description
        ai.Path = actInfo.Path
        ai.Method = actInfo.Method
        ai.Parameters = (service.Parameters)(actInfo.Parameters)

        c.apiInformation = append(c.apiInformation, ai)
    }

    return c.apiInformation
}
