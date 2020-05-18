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

package hotStuff

import (
    "SealABC/metadata/message"
    "SealABC/log"
)

func (b *basicService) sendMessageToLeader(msg message.Message) {
    b.viewChangeTrigger.Reset(b.config.ConsensusTimeout)
    leader := b.getLeader()
    //todo : modular log system
    //log.Log.Println("send message to leader node: ", leader.FromNode, " msg : ", msg.Type)
    go b.network.SendTo(leader.FromNode, msg)
}

func (b *basicService) broadCastMessage(msg message.Message) {
    b.viewChangeTrigger.Reset(b.config.ConsensusTimeout)

    //todo: modular log system
    //log.Log.Println("broadcast message: ", msg.Type)
    for _, m := range b.config.Members {
        if m.Signer.PublicKeyString() == b.config.SelfSigner.PublicKeyString() {
            continue
        }

        n := m.FromNode
        go func() {
            _, err := b.network.SendTo(n, msg)
            if err != nil {
                log.Log.Error("send message to member failed: ", msg.Type, " ",n.ServeAddress)
            }
        }()
    }
}
