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
    "github.com/SealSC/SealABC/consensus"
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/message"
)

func (b *basicService) pickHighQC() (highQC QC) {
    highView := -1
    blankNewViewQC := QC {}

    for _, newView := range b.newViews {
        if highView < int(newView.ViewNumber) {
            highView = int(newView.ViewNumber)
            highQC = newView.Justify
        }
    }

    if highView == 0 {
        for _, newView := range b.newViews {
            blankNewViewQC.Votes = append(blankNewViewQC.Votes, newView.Justify.Votes...)
        }
        highQC.Votes = blankNewViewQC.Votes
    }

    return
}

func (b *BasicService) GotVote(consensusData SignedConsensusData) (reply *message.Message) {
	if !b.hotStuff.GotVoteRule(b, consensusData) {
		return
	}
	if !b.verifyVoteMessage(consensusData) {
		return
	}
	if !b.HasEnoughVotes(len(b.VotedMessage)) {
		return
	}
	b.hotStuff.OnReceiveVote(b, consensusData.ConsensusData)

	return
}

func (b *BasicService) GotNewView(consensusData SignedConsensusData) (_ *message.Message) {
	if consensusData.ViewNumber < b.CurrentView {
		//log.Log.Error("new view is lower then my, do nothing.")
		return
	}

	if !b.verifyNewViewMessage(consensusData) {
		log.Log.Error("not a valid new view.")
		return
	}

	b.hotStuff.NewView(b, consensusData)

	return
}
