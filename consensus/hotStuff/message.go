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
	"encoding/json"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/message"
	"github.com/SealSC/SealABC/metadata/seal"
)

type messageType struct {
	NewView   enum.Element
	Prepare   enum.Element
	PreCommit enum.Element
	Commit    enum.Element
	Decide    enum.Element
	Vote      enum.Element
	Generic   enum.Element
}

var MessageTypes messageType

func (b *BasicService) consensusDataFromMessage(msg message.Message) (consensusData SignedConsensusData, err error) {
	consensusData = SignedConsensusData{}
	err = json.Unmarshal(msg.Payload, &consensusData)
	if err != nil {
		log.Log.Error("invalid consensusData")
	}

	return
}

func (b *BasicService) BuildVote(qcData QCData) (vote seal.Entity, err error) {
	qcBytes, err := structSerializer.ToMFBytes(qcData)
	if err != nil {
		log.Log.Error("serialize QC data failed")
		return
	}

	qcHash := sha3.Sha256.Sum(qcBytes)
	sig, err := b.Config.SelfSigner.Sign(qcHash)
	if err != nil {
		return
	}

	vote = seal.Entity{
		Hash:            qcHash,
		SignerPublicKey: b.Config.SelfSigner.PublicKeyBytes(),
		Signature:       sig,
	}

	return
}

func (b *BasicService) BuildConsensusMessage(phase string, payload ConsensusPayload, justify QC, parentId string, viewNumber uint64) (msgPayload []byte, err error) {
	consensusMsg := SignedConsensusData{}

	consensusMsg.ViewNumber = viewNumber
	consensusMsg.Phase = phase
	consensusMsg.Justify = justify
	consensusMsg.Payload = payload
	consensusMsg.ParentId = parentId
	consensusMsg.Id = consensusMsg.ConsensusData.NodeId()

	dataForSign := QCData{
		Phase:      phase,
		ViewNumber: viewNumber,
		Payload:    payload,
	}

	//serialize to bytes
	consensusBytes, err := structSerializer.ToMFBytes(dataForSign)
	if err != nil {
		log.Log.Error("serialize data failed.")
		return
	}

	//sign
	consensusMsg.Seal.Hash = sha3.Sha256.Sum(consensusBytes)
	consensusMsg.Seal.SignerPublicKey = b.Config.SelfSigner.PublicKeyBytes()
	consensusMsg.Seal.Signature, err = b.Config.SelfSigner.Sign(consensusMsg.Seal.Hash)
	if err != nil {
		return
	}
	//marshal data to json
	msgPayload, err = json.Marshal(consensusMsg)
	if err != nil {
		log.Log.Error("consensus data marshal to json failed.")
		return
	}

	return

func (b *BasicService) BuildCommonMessage(msgType enum.Element, msgPayload []byte) (msg message.Message) {
	msg.Family = b.hotStuff.MessageFamily()
	msg.Version = b.hotStuff.MessageVersion()
	msg.Type = msgType.String()
	msg.Payload = msgPayload

	return
}

func (b *BasicService) buildNewViewQC() (qc *QC) {
	newQC := QC{}
	newQC.Phase = ConsensusPhases.Prepare.String()

	if b.CurrentView == 0 {
		//in genesis phase, members will send a empty qc to leader
		newQC.ViewNumber = 0
	} else {
		if b.PrepareQC != nil {
			newQC = *b.PrepareQC
		} else {
			//the null prepare qc situation should be caused by first consensus round failed or recovered from a crash.
			//set view number to zero
			newQC.ViewNumber = 0
		}
	}

	if newQC.ViewNumber == 0 {
		appendPrevVotes := true
		if len(newQC.Votes) > 0 {
			appendPrevVotes = !b.Config.SelfSigner.PublicKeyCompare(newQC.Votes[0].SignerPublicKey)
		}
		if appendPrevVotes {
			vote, _ := b.BuildVote(newQC.QCData)
			newQC.Votes = append(newQC.Votes, vote)
		}
	}

	return &newQC
}

func (b *BasicService) BuildNewViewMessage() (msg message.Message, err error) {
	newViewQC := b.buildNewViewQC()
	if newViewQC == nil {
		return
	}

	msgPayload, err := b.hotStuff.BuildNewViewMessage(b, *newViewQC)

	if err != nil {
		return
	}

	msg = b.BuildCommonMessage(MessageTypes.NewView, msgPayload)
	return
}

func (b *BasicService) BuildCommonPhaseMessage(
	phase enum.Element,
	msgType enum.Element,
	votedQC *QC) (msg message.Message, err error) {


	for _, votedMsg := range b.VotedMessage {
		votedQC.Votes = append(votedQC.Votes, votedMsg.Seal)
	}

	selfVote, err := b.BuildVote(votedQC.QCData)
	if err != nil {
		return
	}

	votedQC.Votes = append(votedQC.Votes, selfVote)

	msgPayload, err := b.BuildConsensusMessage(
		phase.String(),
		ConsensusPayload{},
		*votedQC,
		votedQC.NodeId,
		b.CurrentView)

	if err != nil {
		return
	}

	msg = b.BuildCommonMessage(msgType, msgPayload)
	return
}

func (b *BasicService) buildLeafNode(basedQC QC) (node ConsensusPayload, err error) {
	customerData, err := b.ExternalProcessor.CustomerDataToConsensus(b.GetLastConsensusCustomerData())
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

	node.Parent = b.Config.HashCalc.Sum(parentNodeBytes)
	return
}

func (b *BasicService) BroadCastPrepareMessage(highQC QC) (err error) {
	payload, err := b.buildLeafNode(highQC)
	if err != nil {
		return
	}

	err = b.hotStuff.OnProposal(b, highQC, payload)

	return
}

func (b *BasicService) BuildVoteMessage(phase string, payload ConsensusPayload, parentId string, viewNumber uint64) (msg message.Message, err error) {
	//todo: rebuild the payload (to extends parallel service)

	//build payload
	msgPayload, err := b.BuildConsensusMessage(
		phase,
		payload,
		QC{},
		parentId,
		viewNumber)

	if err != nil {
		return
	}

	//build message
	msg = b.BuildCommonMessage(MessageTypes.Vote, msgPayload)

	return
}
