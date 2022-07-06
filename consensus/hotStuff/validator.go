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
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/seal"
)

func (b *BasicService) VerifyQCVotes(qc QC) (passed bool) {
	passed = false

	voteCounter := map[string]bool{}

	signedData := qc.QCData
	signedBytes, _ := structSerializer.ToMFBytes(signedData)

	voterCount := len(qc.Votes)
	for _, v := range qc.Votes {
		if !b.isMemberKey(v.SignerPublicKey) {
			log.Log.Error("signature not a from a member!")
			continue
		}

		signer, _ := b.Config.SingerGenerator.FromRawPublicKey(v.SignerPublicKey)
		qcHash := sha3.Sha256.Sum(signedBytes)
		if passed, _ = signer.Verify(qcHash, v.Signature); passed {
			voteCounter[signer.PublicKeyString()] = true
		} else {
			log.Log.Info("invalid signature: ", signer.PublicKeyString())
			log.Log.Error("invalid signature")
		}
	}

	log.Log.Info("--------", voterCount, "-----", len(voteCounter))
	if voterCount != 1 {
		passed = b.HasEnoughVotes(len(voteCounter))
	} else {
		passed = len(voteCounter) == 1
	}

	if !passed {
		log.Log.Error("not enough valid vote: ", voterCount, " counter: ", len(voteCounter))
	}
	return
}

func (b *BasicService) newQCForMatching(consensusData ConsensusData) (qcForMatching *QCData) {
	allPhases := ConsensusPhases
	qcForMatching = &QCData{}

	switch consensusData.Phase {
	case allPhases.PreCommit.String():
		qcForMatching.Phase = allPhases.Prepare.String()
		qcForMatching.ViewNumber = b.CurrentView

	case allPhases.Commit.String():
		qcForMatching.Phase = allPhases.PreCommit.String()
		qcForMatching.ViewNumber = b.CurrentView

	case allPhases.Decide.String():
		qcForMatching.Phase = allPhases.Commit.String()
		qcForMatching.ViewNumber = b.CurrentView
	}

	return
}

func (b *BasicService) verifySignature(data interface{}, sig seal.Entity) (passed bool) {
	passed = false
	dataBytes, err := structSerializer.ToMFBytes(data)
	if err != nil {
		log.Log.Error("serialize data failed.")
		return
	}

	hash := sha3.Sha256.Sum(dataBytes)
	signer, _ := b.Config.SingerGenerator.FromRawPublicKey(sig.SignerPublicKey)
	passed, _ = signer.Verify(hash, sig.Signature)
	return
}

func (b *BasicService) verifyPhase(consensusData ConsensusData) (passed bool) {
	passed = false
	qc := consensusData.Justify
	qcData := qc.QCData

	matchingQC := b.newQCForMatching(consensusData)
	if matchingQC == nil {
		log.Log.Error("can't get a basic QC for matching.")
		return
	}

	if qcData.ViewNumber != matchingQC.ViewNumber {
		if qcData.ViewNumber < matchingQC.ViewNumber {
			log.Log.Error("in phase ", qcData.Phase, " verify found QC not in the same [VIEW]. [local@", matchingQC.ViewNumber, " : remote@", qcData.ViewNumber, "]")
		} else {
			passed = b.VerifyQCVotes(qc)
			if passed {
				log.Log.Warn("in phase ", qcData.Phase, " current view is lower then network, but QC is valid. maybe I am behind")
			}
		}
		return
	}

	if qcData.Phase != matchingQC.Phase {
		log.Log.Error("in phase verify found QC not in the same [PHASE]. [local@", matchingQC.Phase, " : remote@", qcData.Phase, " : msg@"+consensusData.Phase+"]")
		return
	}

	passed = b.VerifyQCVotes(qc)

	return
}

func (b *BasicService) verifyVoteMessage(consensusData SignedConsensusData) (passed bool) {
	passed = false
	//check message phase
	if consensusData.Phase != b.CurrentPhase.String() {
		//log.Log.Error("message not at the same [PHASE]. ", " [local@",
		//    b.currentPhase.String(), " : remote@", consensusData.Phase, "]")
		return
	}

	//check message view
	if consensusData.ViewNumber != b.CurrentView {
		log.Log.Info("message not at the same [VIEW]. ", " [local@",
			b.CurrentView, " : remote@", consensusData.ViewNumber, "]")
		return
	}

	memberKey := consensusData.Seal.SignerPublicKey
	if !b.isMemberKey(memberKey) {
		log.Log.Error("not valid member who sent this vote!")
		return
	}

	singer, _ := b.Config.SingerGenerator.FromRawPublicKey(memberKey)
	singerHexKey := singer.PublicKeyString()
	if _, voted := b.VotedMessage[singerHexKey]; voted {
		return true
	}

	b.VotedMessage[singerHexKey] = consensusData
	return true
}

func (b *BasicService) verifyNewViewMessage(consensusData SignedConsensusData) (passed bool) {
	passed = false

	justify := consensusData.Justify
	if consensusData.ViewNumber < justify.ViewNumber {
		log.Log.Error("invalid consensus view and justify view.")
		return
	}

	if len(justify.Votes) > 0 {
		passed = b.VerifyQCVotes(justify)
	} else if b.CurrentView == 0 {
		passed = true
	}

	return
}

func (b *BasicService) verifyPrepareMessage(consensusData SignedConsensusData) (passed bool) {
	return b.hotStuff.VerifyProposal(b, consensusData)
}

func (b *BasicService) HasEnoughVotes(voteCount int) bool {
	memberCount := len(b.Config.Members)
	bftCount := memberCount / 3

	needCount := memberCount - bftCount - 1 //exclude leader self in enough count.

	return needCount <= voteCount
}
