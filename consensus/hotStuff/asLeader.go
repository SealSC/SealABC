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

func (b *basicService) getNextPhaseAndMsgType() (phase enum.Element, msgType enum.Element) {
    switch b.currentPhase {
    case consensusPhases.NewView:
        phase = consensusPhases.Prepare
        msgType = messageTypes.Prepare

    case consensusPhases.Prepare:
        phase = consensusPhases.PreCommit
        msgType = messageTypes.PreCommit

    case consensusPhases.PreCommit:
        phase = consensusPhases.Commit
        msgType = messageTypes.Commit

    case consensusPhases.Commit:
        phase = consensusPhases.Decide
        msgType = messageTypes.Decide

    case consensusPhases.Decide:
        phase = consensusPhases.NewView
        msgType = messageTypes.NewView
    }

    return
}

func (b *basicService) gotVote(consensusData SignedConsensusData) (reply *message.Message) {
    if !b.verifyVoteMessage(consensusData) {
        return
    }

    if !b.hasEnoughVotes(len(b.votedMessage)) {
        return
    }

    nextPhase, msgType := b.getNextPhaseAndMsgType()

    //todo: pre-commit message need rebuild payload from all votes for parallel service
    votedQC := QC{}
    votedQC.Payload = consensusData.Payload
    votedQC.ViewNumber = b.currentView
    votedQC.Phase = consensusData.Phase

    nextPhaseMsg, err := b.buildCommonPhaseMessage(nextPhase, msgType, &votedQC)
    if err != nil{
        return
    }

    switch b.currentPhase {
    case consensusPhases.PreCommit:
        b.prepareQC = &votedQC

    case consensusPhases.Commit:
        b.lockedQC = &votedQC
    }

    b.currentPhase = nextPhase
    b.votedMessage = map[string]SignedConsensusData{}
    b.broadCastMessage(nextPhaseMsg)

    if nextPhase == consensusPhases.Decide {
        if b.externalProcessor != nil {
            b.externalProcessor.EventProcessor(consensus.Event.Success, consensusData.Payload.CustomerData)
        }

        b.currentPhase = consensusPhases.NewView
        b.currentView += 1

        b.newRound()
    }
    return
}

func (b *basicService) gotNewView(consensusData SignedConsensusData) (_ *message.Message) {
    if consensusData.ViewNumber < b.currentView {
        log.Log.Error("new view is lower then my, do nothing.")
        return
    }

    if !b.verifyNewViewMessage(consensusData) {
        log.Log.Error("not a valid new view.")
        return
    }

    singer, _ := b.config.SingerGenerator.FromRawPublicKey(consensusData.Seal.SignerPublicKey)
    b.newViews[singer.PublicKeyString()] = consensusData

    if !b.hasEnoughVotes(len(b.newViews)) {
        return
    }

    //if consensusData.ViewNumber == b.currentView && b.currentPhase != consensusPhases.NewView {
    //    return
    //}

    b.currentView = consensusData.ViewNumber
    highQC := b.pickHighQC()

    prepareMsg, err := b.buildPrepareMessage(highQC)
    if err != nil {
        log.Log.Error("build prepare message failed.")
        b.newViews = map[string]SignedConsensusData{}
        return
    }

    b.newViews = map[string]SignedConsensusData{}
    b.currentPhase = consensusPhases.Prepare
    b.broadCastMessage(prepareMsg)
    return
}

func (b *basicService) registerLeaderProcessor() {
    b.consensusProcessor[messageTypes.NewView.String()] = b.gotNewView
    b.consensusProcessor[messageTypes.Vote.String()] = b.gotVote
}
