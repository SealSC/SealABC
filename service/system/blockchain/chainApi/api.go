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

package chainApi

import (
    "SealABC/service/system/blockchain/chainApi/httpJSON"
    "SealABC/service/system/blockchain/chainApi/httpJSON/actions"
    "SealABC/service/system/blockchain/chainSQLStorage"
    "SealABC/service/system/blockchain/chainStructure"
    "SealABC/dataStructure/enum"
    "SealABC/service/system/blockchain/chainNetwork"
)

type ApiServers struct {
    HttpJSON        *httpJSON.ApiServer
}

func Load()  {
    enum.SimpleBuild(&actions.URLParameterKeys)
}

func NewServer(cfg Config, chain *chainStructure.Blockchain, p2p *chainNetwork.P2PService, sqlStorage *chainSQLStorage.Storage) *ApiServers {
    api := ApiServers{}

    api.HttpJSON = httpJSON.NewApiServer(cfg.HttpJSON, chain, p2p, sqlStorage)

    return &api
}
