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
    "net"
    "io"
    "bufio"
    "github.com/SealSC/SealABC/log"
    "sync"
)

const (
    max_buffer_size  = 10 * 1024 * 1024
)

type ILink interface {
    Start()
    StartReceiving()
    SendData(data [] byte) (n int, err error)
    SendMessage(msg Message) (n int, err error)
    RemoteAddr() net.Addr
    Close()
}

type Link struct {
    Connection          net.Conn
    Reader              *bufio.Reader
    Writer              *bufio.Writer
    RawMessageProcessor RawMessageProcessor
    LinkClosed          LinkClosed
    ConnectOut          bool

    senderLock          sync.Mutex
}

func (l *Link) RemoteAddr() net.Addr {
    return l.Connection.RemoteAddr()
}

func (l *Link)Start() {

    l.Reader = bufio.NewReader(l.Connection)
    l.Writer = bufio.NewWriter(l.Connection)

    go l.StartReceiving()
}

func (l *Link)StartReceiving() {
    defer func() {
        l.Close()
        l.LinkClosed(l)
    }()

    for {
        msgPrefix := make([]byte, MESSAGE_PREFIX_LEN, MESSAGE_PREFIX_LEN)
        data := make([]byte, max_buffer_size, max_buffer_size)

        n, err := io.ReadFull(l.Reader, msgPrefix) //l.Reader.Read(msgPrefix[:])
        if n != MESSAGE_PREFIX_LEN {
            log.Log.Println("get msg prefix failed: need ", MESSAGE_PREFIX_LEN, " bytes, got ", n, " bytes. ", err)
        }

        if err != nil {
            if err == io.EOF {
                log.Log.Println("disconnect remote: ", err)
                break
            }
            log.Log.Println("got an network error: ", err)
            continue
        }

        prefix := MessagePrefix{}
        err = prefix.FromBytes(msgPrefix)
        if err != nil {
            log.Log.Warn("unknown message: ", err.Error())
            unreadSize := l.Reader.Size()
            _, _ =l.Reader.Discard(unreadSize)
            continue
        }

        if prefix.Size > MAX_MESSAGE_LEN {
            _, _ = l.Reader.Discard(int(prefix.Size))
            continue
        }

        n, err = io.ReadFull(l.Reader, data[:prefix.Size])
        //n, err := l.Reader.Read(data[:prefix.Size])

        if int32(n) != prefix.Size {
            log.Log.Println("error message: need ", prefix.Size, "bytes bug got ", n, " bytes")
            //log.Log.Println("error ", err.Error())
            return
        }

        if err != nil {
            if err == io.EOF {
                log.Log.Println("disconnect remote: ", err)
                break
            }
            log.Log.Println("got an network error: ", err)
            continue
        }

        go l.RawMessageProcessor(data[:prefix.Size], l)
    }
}

func (l *Link)SendMessage(msg Message) (n int, err error) {
    data, err := msg.ToRawMessage()
    if err != nil {
        return
    }

    return l.SendData(data)
}

func (l *Link)SendData(data []byte) (n int, err error) {
    l.senderLock.Lock()
    defer l.senderLock.Unlock()

    n, err = l.Writer.Write(data)
    if err != nil {
        log.Log.Warn("got an error when write to buffer: ", err.Error())
        return
    }

    if n != len(data) {
        log.Log.Warn("not write the complete data to buffer: ", n, " bytes written but need: ", len(data), " bytes")
    }

    err = l.Writer.Flush()

    if n == 0 {
        log.Log.Println("sent 0 bytes to ", l.RemoteAddr())
        if err != nil {
            log.Log.Warn(" and got an error: ", err.Error())
        }
    }
    return
}

func (l *Link)Close() {
    l.Connection.Close()
}
