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
    "SealABC/common/utility/serializer/structSerializer"
    "SealABC/crypto/hashes/sha3"
    "SealABC/dataStructure/enum"
    "SealABC/log"
    "SealABC/metadata/message"
    "SealABC/metadata/seal"
    "encoding/json"
)

const MessageFamily = "basic-hot-stuff-consensus"
const MessageVersion = "basic.0.0.1"

type messageType struct {
    NewView     enum.Element
    Prepare     enum.Element
    PreCommit   enum.Element
    Commit      enum.Element
    Decide      enum.Element
    Vote        enum.Element

}

var messageTypes messageType

func (b *basicService) consensusDataFromMessage(msg message.Message) (consensusData SignedConsensusData, err error)  {
    consensusData = SignedConsensusData{}
    err = json.Unmarshal(msg.Payload, &consensusData)
    if err != nil {
        log.Log.Error("invalid consensusData")
    }

    return
}

func (b *basicService) buildVote(payload ConsensusPayload) (vote seal.Entity, err error) {
    qcBytes, err := structSerializer.ToMFBytes(payload)
    if err != nil {
        log.Log.Error("serialize QC data failed")
        return
    }

    qcHash := sha3.Sha256.Sum(qcBytes)
    sig := b.config.SelfSigner.Sign(qcHash)

    vote = seal.Entity {
        Hash: qcHash,
        SignerPublicKey: b.config.SelfSigner.PublicKeyBytes(),
        Signature:       sig,
    }

    return
}

func (b *basicService) buildConsensusMessage(phase string, payload ConsensusPayload, justify QC) (msgPayload []byte, err error) {
    consensusMsg := SignedConsensusData{}

    consensusMsg.ViewNumber = b.currentView
    consensusMsg.Phase = phase
    consensusMsg.Justify = justify
    consensusMsg.Payload = payload

    dataForSign := QCData {
        Phase: phase,
        ViewNumber: b.currentView,
        Payload: payload,
    }

    //serialize to bytes
    consensusBytes, err := structSerializer.ToMFBytes(dataForSign)
    if err != nil {
        log.Log.Error("serialize data failed.")
        return
    }

    //sign
    consensusMsg.Seal.Hash = sha3.Sha256.Sum(consensusBytes)
    consensusMsg.Seal.SignerPublicKey = b.config.SelfSigner.PublicKeyBytes()
    consensusMsg.Seal.Signature = b.config.SelfSigner.Sign(consensusMsg.Seal.Hash)

    //marshal data to json
    msgPayload, err = json.Marshal(consensusMsg)
    if err != nil {
        log.Log.Error("consensus data marshal to json failed.")
        return
    }

    return
}

func (b *basicService) buildCommonMessage(msgType enum.Element, msgPayload []byte) (msg message.Message) {
    msg.Family = MessageFamily
    msg.Version = MessageVersion
    msg.Type = msgType.String()
    msg.Payload = msgPayload

    return
}

func (b *basicService) buildNewViewQC() (qc *QC) {
    newQC := QC{}
    newQC.Phase =  consensusPhases.Prepare.String()

    if b.currentView == 0 {
        //in genesis phase, members will send a empty qc to leader
        newQC.ViewNumber = 0
    } else {
        if b.prepareQC != nil {
            newQC = *b.prepareQC
        } else {
            //the null prepare qc situation should be caused by first consensus round failed or recovered from a crash.
            //set view number to zero
            newQC.ViewNumber = 0
        }
    }

    if newQC.ViewNumber == 0 {
        appendPrevVotes := true
        if len(newQC.Votes) > 0 {
            appendPrevVotes = !b.config.SelfSigner.PublicKeyCompare(newQC.Votes[0].SignerPublicKey)
        }
        if appendPrevVotes {
            vote, _ := b.buildVote(newQC.Payload)
            newQC.Votes = append(newQC.Votes, vote)
        }
    }

    return &newQC
}

func (b *basicService) buildNewViewMessage() (msg message.Message, err error) {
    newViewQC := b.buildNewViewQC()
    if newViewQC == nil {
        return
    }

    msgPayload, err := b.buildConsensusMessage(
        consensusPhases.NewView.String(),
        ConsensusPayload{},
        *newViewQC)

    if err != nil {
        return
    }

    msg = b.buildCommonMessage(messageTypes.NewView, msgPayload)
    return
}

func (b *basicService) buildCommonPhaseMessage(
    phase enum.Element,
    msgType enum.Element,
    votedQC *QC) (msg message.Message, err error) {

    for _, votedMsg := range b.votedMessage {
        votedQC.Votes = append(votedQC.Votes, votedMsg.Justify.Votes...)
    }

    selfVote, err := b.buildVote(votedQC.Payload)
    if err != nil {
        return
    }

    votedQC.Votes = append(votedQC.Votes, selfVote)

    msgPayload, err := b.buildConsensusMessage(
        phase.String(),
        ConsensusPayload{},
        *votedQC)

    if err != nil {
        return
    }

    msg = b.buildCommonMessage(msgType, msgPayload)
    return
}


func (b *basicService) buildLeafNode(basedQC QC) (node ConsensusPayload, err error) {
    customerData, err := b.externalProcessor.CustomerDataToConsensus()
    if err != nil {
        log.Log.Error("get customer data failed.")
        return
    }

    parentNodeBytes, _ := structSerializer.ToMFBytes(basedQC.Payload)

    node.CustomerData, err = customerData.Bytes()
    if err != nil {
        log.Log.Error("customer data builder returns an error: ", err)
        return
    }

    node.Parent = b.config.HashCalc.Sum(parentNodeBytes)
    return
}

func (b *basicService) buildPrepareMessage(highQC QC) (msg message.Message, err error) {
    payload, err := b.buildLeafNode(highQC)
    if err != nil {
        return
    }

    msgPayload, err := b.buildConsensusMessage(
        consensusPhases.Prepare.String(),
        payload,
        highQC)

    if err != nil {
        return
    }

    //build message
    msg = b.buildCommonMessage(messageTypes.Prepare, msgPayload)
    return
}

func (b *basicService) buildVoteMessage(phase string, payload ConsensusPayload) (msg message.Message, err error) {
    //todo: rebuild the payload (to extends parallel service)

    //build payload
    msgPayload, err := b.buildConsensusMessage(
        phase,
        payload,
        QC{})

    if err != nil {
        return
    }

    //build message
    msg = b.buildCommonMessage(messageTypes.Vote, msgPayload)

    return
}
