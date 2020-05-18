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
    "sync"
    "SealABC/network"
    "time"
    "SealABC/log"
)

var syncBlockWait sync.WaitGroup
var Syncing = false

func (p *P2PService) StartSync(nodes []network.Node, targetHeight uint64) {
    if Syncing {
        return
    }

    p.syncLock.Lock()
    Syncing = true

    defer func() {
        if r := recover(); r != nil {
            log.Log.Error("got a panic: ", r)
        }

        Syncing = false
        p.syncLock.Unlock()
    }()

    seedsCnt := len(nodes)
    for s := p.chain.CurrentHeight(); s < targetHeight; s++ {
        var syncErr error = nil
        seedIdx := 0
        for i := 0; i< seedsCnt; i++ {
            seedIdx = (int(s) + 1) % seedsCnt
            syncErr = p.syncBlockFrom(nodes[seedIdx], s + 1)

            if syncErr == nil {
                syncBlockWait.Add(1)
                break
            }
        }

        if syncErr != nil {
            log.Log.Error("sync block ", s, " failed")
            time.Sleep(time.Second)
            continue
        }

        log.Log.Println("waiting for block ",  s + 1, " from node ", nodes[seedIdx].ServeAddress)
        syncBlockWait.Wait()
        log.Log.Println("sync block ",  s + 1, " over.")
    }
    return
}

func (p *P2PService) syncBlockFrom(node network.Node, height uint64) (err error) {
    reqMsg := newSyncBlockMessage(height)
    p.NetworkService.SendTo(node, reqMsg)
    return
}

