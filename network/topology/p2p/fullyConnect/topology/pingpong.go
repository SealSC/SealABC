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

package topology

import (
    "encoding/json"
    "github.com/SealSC/SealABC/network"
    "github.com/SealSC/SealABC/network/topology/p2p/fullyConnect/message"
    "github.com/SealSC/SealABC/network/topology/p2p/fullyConnect/message/payload"
    "time"
)

func doPing(link network.ILink) {
    ping := payload.NewPing()
    pingPayloadBytes, _ := json.Marshal(ping)
    msg := message.NewMessage(message.Types.Ping, pingPayloadBytes)
    link.SendMessage(msg)
}

type pingMessageProcessor struct {}

func (p *pingMessageProcessor)Process(msg network.Message, t *Topology, link network.ILink)  (err error) {
    pingPayload := payload.PingPongPayload{}
    err = payload.FromMessage(msg, &pingPayload)
    if err != nil {
        return
    }

    pong := payload.PingPongPayload{
        Number: pingPayload.Number,
    }

    replayPayloadBytes, _ := json.Marshal(pong)
    reply := message.NewMessage(message.Types.Pong, replayPayloadBytes)
    link.SendMessage(reply)

    go func() {
        time.Sleep(time.Second * 10)

        doPing(link)
    }()


    return
}

type pongMessageProcessor struct {}
func (p *pongMessageProcessor)Process(msg network.Message, topology *Topology, _ network.ILink)  (err error) {
    //todo: refresh neighbor's state

    return
}


var PingMessageProcessor = &pingMessageProcessor{}
var PongMessageProcessor = &pongMessageProcessor{}
