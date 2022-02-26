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
    "github.com/SealSC/SealABC/engine/engineApi"
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/network"
    "github.com/SealSC/SealABC/service/system"
)

type Config struct {
    Api         engineApi.Config
    Log         log.Config

    StorageConfig           interface{}
    ConsensusNetwork        network.Config
    ConsensusDisabled       bool

    Consensus        interface{}
    SystemService   system.Config

}

var config Config

func loadConfig(cfg Config) {
    config = cfg
    return
}
