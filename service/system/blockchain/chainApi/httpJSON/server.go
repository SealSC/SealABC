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

package httpJSON

import (
    "SealABC/network/http"
    "SealABC/service/system/blockchain/chainApi/httpJSON/actions"
    "SealABC/service/system/blockchain/chainSQLStorage"
    "SealABC/service/system/blockchain/chainStructure"
    "SealABC/service/system/blockchain/chainNetwork"
)

type ApiServer struct {
    server  http.Server
    Actions *actions.ChainApiActions
}

func (a ApiServer) Address() string {
    return a.server.Config.Address
}

func NewApiServer(cfg http.Config,
    chain *chainStructure.Blockchain,
    p2p *chainNetwork.P2PService,
    sqlStorage *chainSQLStorage.Storage) *ApiServer {

    httpServer := http.Server {
        Config: &cfg,
    }

    httpServer.Config.AllowCORS = true

    newActions := actions.NewActions(cfg.BasePath, chain, p2p, sqlStorage)
    httpServer.Config.RequestHandler = []http.IRequestHandler{newActions}

    _ = httpServer.Start()

    as := ApiServer {
        server: httpServer,
        Actions: newActions,
    }

    return &as
}
