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
	"github.com/SealSC/SealABC/metadata/message"
)

func (b *BasicService) SendMessageToLeader(msg message.Message) {
	b.ViewChangeTrigger.Reset(b.Config.ConsensusTimeout)
	leader := b.getLeader()
	//todo : modular log system
	//log.Log.Println("send message to leader node: ", leader.FromNode, " msg : ", msg.Type)
	go b.network.SendTo(leader.FromNode, msg)
}

func (b *BasicService) BroadCastMessage(msg message.Message) {
	b.ViewChangeTrigger.Reset(b.Config.ConsensusTimeout)

	//todo: modular log system
	//log.Log.Println("broadcast message: ", msg.Type)
	go b.network.Broadcast(msg)
}
