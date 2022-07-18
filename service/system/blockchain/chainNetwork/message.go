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
	"encoding/json"
	"errors"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/metadata/message"
)

const messageFamily = "seal-chain-message"
const messageVersion = "0.1"

var MessageTypes struct {
	PushRequest    enum.Element
	SyncBlock      enum.Element
	SyncBlockReply enum.Element
}

type syncBlockReplyMessage struct {
	Success bool
	Block   block.Entity
}

type syncBlockMessage struct {
	BlockHeight uint64
}

func getBlockFromSyncReplyMessage(msg message.Message) (blk *block.Entity, err error) {
	replyMsg := syncBlockReplyMessage{}
	err = json.Unmarshal(msg.Payload, &replyMsg)
	if err != nil {
		return
	}

	if !replyMsg.Success {
		err = errors.New("block not exits")
		return
	}

	blk = &replyMsg.Block
	return
}

func getHeightFromSyncMessage(msg message.Message) (height uint64, err error) {
	syncBlkMsg := syncBlockMessage{}
	err = json.Unmarshal(msg.Payload, &syncBlkMsg)
	if err != nil {
		return
	}

	height = syncBlkMsg.BlockHeight
	return
}

func getBlockchainRequestFromPushRequestMessage(msg message.Message) (req blockchainRequest.Entity, err error) {
	err = json.Unmarshal(msg.Payload, &req)
	if err != nil {
		return
	}

	return
}

func newSyncBlockMessage(height uint64) (msg message.Message) {
	syncMsg := syncBlockMessage{
		BlockHeight: height,
	}

	payload, _ := json.Marshal(syncMsg)
	msg = newMessage(MessageTypes.SyncBlock, payload)
	return
}

func newSyncBlockReplyMessage(blk *block.Entity, height uint64) (msg message.Message) {
	replyMsg := syncBlockReplyMessage{}

	if blk == nil {
		replyMsg.Success = false
		replyMsg.Block.Header.Height = height
	} else {
		replyMsg.Success = true
		replyMsg.Block = *blk
	}

	payload, _ := json.Marshal(replyMsg)
	msg = newMessage(MessageTypes.SyncBlockReply, payload)

	return
}

func NewPushRequest(req blockchainRequest.Entity) (msg message.Message, err error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return
	}

	msg = newMessage(MessageTypes.PushRequest, payload)
	return
}

func newMessage(msgType enum.Element, payload []byte) (msg message.Message) {
	msg.Family = messageFamily
	msg.Version = messageVersion
	msg.Type = msgType.String()
	msg.Payload = payload
	return
}
