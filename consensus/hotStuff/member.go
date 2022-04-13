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
	"github.com/SealSC/SealABC/crypto/signers/signerCommon"
	"github.com/SealSC/SealABC/network"
)

type Member struct {
	Signer   signerCommon.ISigner
	FromNode network.Node
	online   bool
}

func (b *BasicService) isMemberKey(memberKey []byte) bool {
	ret := false
	for _, m := range b.Config.Members {
		ret = m.Signer.PublicKeyCompare(memberKey)
		if ret {
			break
		}
	}
	return ret
}

func (b *BasicService) isAllMembersOnline() bool {
	allNodes := b.network.GetAllLinkedNode()

	memberCount := len(b.Config.Members) - 1 //exclude self from member count
	onlineCount := map[string]bool{}

	memberMap := map[string]int{}
	for idx, m := range b.Config.Members {
		memberMap[m.Signer.PublicKeyString()] = idx
	}

	for _, n := range allNodes {
		if idx, exist := memberMap[n.ID]; exist {
			b.Config.Members[idx].FromNode = n
			b.Config.Members[idx].online = true

			onlineCount[n.ID] = true
		}
	}

	return memberCount <= len(onlineCount)
}

func (b *BasicService) allMembersKey() (keys []string) {
	for _, m := range b.Config.Members {
		keys = append(keys, m.Signer.PublicKeyString())
	}

	return
}
