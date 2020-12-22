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

package chainNetwork

import (
    "github.com/SealSC/SealABC/network"
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/blockchainRequest"
)

type p2pMessageHandler func(msg network.Message) *network.Message

func (p *P2PService)handlePushRequest(msg network.Message) (_ *network.Message) {
    req, err := getBlockchainRequestFromPushRequestMessage(msg.Message)
    if err != nil {
        log.Log.Error("invalid request from p2p network: ", err.Error())
        return
    }
    _, err = p.chain.Executor.PushRequest(req)
    if err != nil {
        log.Log.Error("push request from p2p network failed: ", err.Error())
        return
    }
    return
}

func (p *P2PService)handleSyncBlock(msg network.Message) (reply *network.Message) {
    height, err := getHeightFromSyncMessage(msg.Message)
    if err != nil {
        log.Log.Error(err.Error())
        replyMsg := newSyncBlockReplyMessage(nil, height)
        reply = &network.Message{
            Message: replyMsg,
        }

        return reply
    }
    blk, err := p.chain.GetBlockByHeight(height)
    if err != nil {
        log.Log.Error("get block@", height, " failed")
        replyMsg := newSyncBlockReplyMessage(nil, height)
        reply = &network.Message{
            Message: replyMsg,
        }
    }

    log.Log.Warn("block@", height, " sync to remote: ", msg.From.ServeAddress)
    replyMsg := newSyncBlockReplyMessage(&blk, height)
    reply = &network.Message{
        Message: replyMsg,
    }

    return
}

func  (p *P2PService)handleSyncBlockReply(msg network.Message) (_ *network.Message) {
    defer func() {
        if r := recover(); r != nil {
            log.Log.Error("got a panic: ", r)
        }
    }()

    blk, err := getBlockFromSyncReplyMessage(msg.Message)
    if err != nil {
        log.Log.Error("get block failed: ", err.Error())
    } else {
        p.chain.AddBlock(*blk)
    }

    //todo: call block sync module's method, not call syncBlockWait directly
    syncBlockWait.Done()
    return
}

func (p *P2PService)handleP2PMessage(msg network.Message) (reply *network.Message) {
    if h, exists := p.networkMessageHandler[msg.Type]; exists {
        return h(msg)
    }
    return
}

func (p *P2PService) BroadcastRequest(req blockchainRequest.Entity) (err error) {
    msg, err := NewPushRequest(req)
    if err != nil {
        return
    }

    err = p.NetworkService.Broadcast(msg)
    return
}
