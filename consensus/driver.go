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

package consensus

import (
    "github.com/SealSC/SealABC/network"
)

type driver struct {
    service IConsensusService
}

var Driver driver

func (d *driver) Start(cfg interface{}) (err error) {
    return d.service.Start(cfg)
}

func (d *driver)messageProcessor(msg network.Message) (reply *network.Message) {
    replyMsg := d.service.Feed(msg.Message)
    if nil != replyMsg {
        reply = &network.Message{
            Message: *replyMsg,
        }
    }
    return
}

func (d *driver)consensusRegister(consensusService IConsensusService, networkService network.IService) {
    d.service = consensusService
    networkService.RegisterMessageProcessor(d.service.GetMessageFamily(), d.messageProcessor)
}
