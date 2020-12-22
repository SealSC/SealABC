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

package network

import (
    "encoding/binary"
    "encoding/json"
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/message"
    "errors"
)



const (
    MSG_JOIN = 0
    MSG_LEAVE = 1
    MSG_BROADCAST = 2
    MSG_MULTICAST = 3
    MSG_UNICAST = 4

    MAGIC_WORD = "[what a day ^_T]"

    MAGIC_WORD_LEN = len(MAGIC_WORD)
    SIZE_LEN       = 4
    SIZE_DATA_TYPE = 1

    MESSAGE_PREFIX_LEN = SIZE_LEN + SIZE_DATA_TYPE + MAGIC_WORD_LEN

    MAX_MESSAGE_LEN = 8 * 1024 * 1024 //raw message max size will be (8 MB + MESSAGE_PREFIX_LEN) bytes.

    JSON_TYPE = 0x00
)

type Message struct {
    message.Message
    From      Node
}

func (m Message) ToRawMessage() (rawMsg []byte, err error)  {
    msgJson, err := json.Marshal(m)
    if err != nil {
        log.Log.Println("marshal message faild: ", err)
        return
    }

    msgSize := len(msgJson)


    rawMsg = append([]byte(MAGIC_WORD))
    rawMsg = append(rawMsg, JSON_TYPE)

    msgSizeBytes := make([]byte, SIZE_LEN, SIZE_LEN)
    binary.BigEndian.PutUint32(msgSizeBytes, uint32(msgSize))
    rawMsg = append(rawMsg, msgSizeBytes...)
    rawMsg = append(rawMsg, msgJson...)

    return
}

func (m *Message) FromRawMessage(msgData []byte) (err error) {
    err = json.Unmarshal(msgData, m)
    if err != nil {
        log.Log.Println("not json message: ", string(msgData))
        return
    }

    return
}

type MessagePrefix struct {
    Size      int32
    DataType  byte
}

func (m *MessagePrefix) FromBytes(prefix []byte) (err error) {
    off := 0
    magic := prefix[ : MAGIC_WORD_LEN]
    if string(magic) != MAGIC_WORD {
        err = errors.New("unknown raw message")
        return
    }
    off += MAGIC_WORD_LEN

    dataType := prefix[off]
    off += SIZE_DATA_TYPE

    if dataType != JSON_TYPE {
        err = errors.New("only support json for now")
        return
    }

    size := binary.BigEndian.Uint32(prefix[off : off + SIZE_LEN])
    if size > MAX_MESSAGE_LEN {
        //err = errors.New("message too large")
        log.Log.Warn("only support ", MAX_MESSAGE_LEN, " bytes message, but got ", size)
    }

    m.Size = int32(size)
    m.DataType = dataType
    return
}
