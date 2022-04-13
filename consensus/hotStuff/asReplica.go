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
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/message"
	"time"
)

func (b *BasicService) GotPrepare(consensusData SignedConsensusData) (reply *message.Message) {
	if !b.verifyPrepareMessage(consensusData) {
		return
	}

	b.hotStuff.OnProposal(b, consensusData)

	return
}

func (b *BasicService) GotGeneric(consensusData SignedConsensusData) (reply *message.Message) {
	if !b.verifyPrepareMessage(consensusData) {
		return
	}

	b.hotStuff.OnProposal(b, consensusData)

	return
}

func (b *BasicService) processCommonPhaseMessage(consensusData ConsensusData) {
	allPhases := ConsensusPhases
	switch consensusData.Phase {
	case allPhases.PreCommit.String():
		b.PrepareQC = &consensusData.Justify
		b.CurrentPhase = allPhases.PreCommit

	case allPhases.Commit.String():
		b.LockedQC = &consensusData.Justify
		b.CurrentPhase = allPhases.Commit

	case allPhases.Decide.String():
		b.CurrentPhase = allPhases.Decide
	}
}

func (b *BasicService) GotCommonPhaseMessage(consensusData SignedConsensusData) (reply *message.Message) {
	validPhase := b.verifyPhase(consensusData.ConsensusData)
	if !validPhase {
		return
	}

	b.processCommonPhaseMessage(consensusData.ConsensusData)

	if b.CurrentPhase == ConsensusPhases.Decide {
		if b.ExternalProcessor != nil {
			b.ExternalProcessor.EventProcessor(consensus.Event.Success, consensusData.Justify.Payload.CustomerData)
		}

		b.CurrentView += 1
		//log.Log.Println("consensus success! need send new view to next leader @view ", b.currentView)
		b.ViewChangeTrigger.Reset(b.Config.ConsensusTimeout)

		newView := b.CurrentView
		go func() {
			time.Sleep(b.Config.ConsensusInterval)
			b.PhaseLock.Lock()
			defer b.PhaseLock.Unlock()
			if b.CurrentView != newView {
				return
			}
			b.NewRound()
		}()

		return
	}

	//log.Log.Println("common phase verify success, start build vote message in phase ", consensusData.Phase)
	voteMsg, err := b.BuildVoteMessage(consensusData.Phase, consensusData.Justify.Payload, consensusData.Id, b.CurrentView)
	if err != nil {
		log.Log.Error("build vote message failed")
		return
	}
	b.ViewChangeTrigger.Reset(b.Config.ConsensusTimeout)

	//log.Log.Println("build vote message in phase ", consensusData.Phase, " over")

	b.SendMessageToLeader(voteMsg)
	return
}
