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
    "github.com/SealSC/SealABC/network"
    "github.com/SealSC/SealABC/network/topology/p2p/fullyConnect/message"
    "github.com/SealSC/SealABC/network/topology/p2p/fullyConnect/message/payload"
    "encoding/json"
    "github.com/SealSC/SealABC/log"
    "sync"
)

func getNeighbors(seed network.LinkNode) {
    log.Log.Println("get neighbors from: ", seed.ID)
    msg := message.NewMessage(message.Types.GetNeighbors, []byte{})
    _, err := seed.Link.SendMessage(msg)
    if err != nil {
        log.Log.Error("send get neighbors message failed: ", err.Error())
    }
}


type getNeighborsMessageProcessor struct {}
func (j *getNeighborsMessageProcessor)Process(msg network.Message, t *Topology, link network.ILink)  (err error) {
    allNodes := t.GetAllNodes()
    neighbors := payload.NeighborsPayload {}

    for _, n := range allNodes {
        neighbors.Neighbors = append(neighbors.Neighbors, n.Node)
    }

    replyPayload, _ := json.Marshal(neighbors)
    replyMsg := message.NewMessage(message.Types.GetNeighborsReply, replyPayload)

    _, err = link.SendMessage(replyMsg)

    return
}

var GetNeighborsMessageProcessor = &getNeighborsMessageProcessor{}

type getNeighborsReplyMessageProcessor struct {
    joinLock sync.Mutex
}
func (j *getNeighborsReplyMessageProcessor)Process(msg network.Message, topology *Topology, link network.ILink)  (err error) {
    neighbors := payload.NeighborsPayload{}
    err = payload.FromMessage(msg, & neighbors)
    if err != nil {
        log.Log.Warn("invalid get neighbor reply message: ", err.Error())
        return err
    }

    for _, n := range neighbors.Neighbors {
        if topology.isJoined(n.ID) {
            continue
        }

        j.joinLock.Lock()
        err := topology.router.JoinTopology(n)
        j.joinLock.Unlock()

        if err != nil {
            log.Log.Warn("join neighbor: ", n, " failed.")
            continue
        }
    }

    return
}

var GetNeighborsReplyMessageProcessor = &getNeighborsReplyMessageProcessor{}
