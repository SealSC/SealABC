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

package engineStartup

import (
    "SealABC/log"
    "SealABC/consensus/hotStuff"
    "SealABC/engine/engineApi"
    "SealABC/engine/engineService"
    //"SealABC/service/system"
)

func Start(cfg Config) (err error) {
    //load config first
    loadConfig(cfg)

    //set up logger for the system
    if log.Log == nil {
        log.SetUpLogger(config.Log)
    }

    //load storage
    loadStorage()

    //start consensus network
    consensusNetwork, _ := startConsensusNetwork()

    //start system service
    startSystemService()

    engineApi.Start(config.Api)


    //start consensus
    //todo: load consensus by config
    err = startConsensus(&hotStuff.Basic, consensusNetwork, &engineService.ConsensusProcessor)
    if err != nil {
        return
    }

    engineService.SetConsensusInformation(&hotStuff.Basic)
    return
}
