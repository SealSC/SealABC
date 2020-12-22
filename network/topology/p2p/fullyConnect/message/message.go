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

package message

import (
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/network"
)

type messageTypes struct {
    Join                enum.Element
    JoinReply           enum.Element
    Leave               enum.Element
    Ping                enum.Element
    Pong                enum.Element
    GetNeighbors        enum.Element
    GetNeighborsReply   enum.Element
    Neighbors           enum.Element
    PushNeighbors       enum.Element

}

const Family = "full-connected-p2p"

var Types messageTypes

func LoadMessageTypes() {
    enum.Build(&Types, 0, "p2p-protocol-msg-")
}

func NewMessage(msgType enum.Element, payload []byte) (msg network.Message) {
    //msg :=  network.Message{}

    msg.Version = "0.0.1"
    msg.Family = Family
    msg.Type = msgType.String()
    msg.Payload = payload

    return
}

