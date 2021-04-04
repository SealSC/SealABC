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
    "bytes"
)

func (b *basicService) verifyQCVotes(qc QC) (passed bool)  {
    passed = false

    payload := qc.Payload
    payloadBytes, _ := structSerializer.ToMFBytes(payload)
    voteCounter := map[string] bool{}

    voterCount := len(qc.Votes)
    for _, v := range qc.Votes {
        if !b.isMemberKey(v.SignerPublicKey) {
            log.Log.Error("signature not a from a member!")
            continue
        }

        signer, _ := b.config.SingerGenerator.FromRawPublicKey(v.SignerPublicKey)
        qcHash := sha3.Sha256.Sum(payloadBytes)
        if passed, _ =signer.Verify(qcHash, v.Signature); passed {
            voteCounter[signer.PublicKeyString()] = true
        } else {
            log.Log.Error("invalid signature")
        }
    }

    if voterCount != 1 {
        passed = b.hasEnoughVotes(len(voteCounter))
    } else {
        passed = len(voteCounter) == 1
    }

    if !passed {
        log.Log.Error("not enough valid vote: ", voterCount, " counter: ", len(voteCounter))
    }
    return
}

func (b *basicService) newQCForMatching(consensusData ConsensusData) (qcForMatching *QCData ) {
    allPhases :=  consensusPhases
    qcForMatching = &QCData{}

    switch consensusData.Phase {
    case allPhases.PreCommit.String():
        qcForMatching.Phase = allPhases.Prepare.String()
        qcForMatching.ViewNumber = b.currentView

    case allPhases.Commit.String():
        qcForMatching.Phase = allPhases.PreCommit.String()
        qcForMatching.ViewNumber = b.currentView

    case allPhases.Decide.String():
        qcForMatching.Phase = allPhases.Commit.String()
        qcForMatching.ViewNumber = b.currentView
    }

    return
}

func (b *basicService) verifySignature(data interface{}, sig seal.Entity) (passed bool) {
    passed = false
    dataBytes, err := structSerializer.ToMFBytes(data)
    if err != nil {
        log.Log.Error("serialize data failed.")
        return
    }

    hash := sha3.Sha256.Sum(dataBytes)
    signer, _ := b.config.SingerGenerator.FromRawPublicKey(sig.SignerPublicKey)
    passed, _ = signer.Verify(hash, sig.Signature)
    return
}

func (b *basicService) verifyPhase(consensusData ConsensusData) (passed bool) {
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
            passed = b.verifyQCVotes(qc)
            if passed {
                log.Log.Warn("in phase ", qcData.Phase, " current view is lower then network, but QC is valid. maybe I am behind")
            }
        }
        return
    }

    if qcData.Phase != matchingQC.Phase {
        log.Log.Error("in phase verify found QC not in the same [PHASE]. [local@", matchingQC.Phase, " : remote@", qcData.Phase, " : msg@" + consensusData.Phase + "]")
        return
    }

    passed = b.verifyQCVotes(qc)

    return
}

func (b *basicService) verifyVoteMessage(consensusData SignedConsensusData) (passed bool) {
    passed = false

    //check message phase
    if consensusData.Phase != b.currentPhase.String() {
        //log.Log.Error("message not at the same [PHASE]. ", " [local@",
        //    b.currentPhase.String(), " : remote@", consensusData.Phase, "]")
        return
    }

    //check message view
    if consensusData.ViewNumber != b.currentView {
        log.Log.Error("message not at the same [VIEW]. ", " [local@",
            b.currentView, " : remote@", consensusData.ViewNumber, "]")
        return
    }

    memberKey := consensusData.Seal.SignerPublicKey
    if !b.isMemberKey(memberKey) {
        log.Log.Error("not valid member who sent this vote!")
        return
    }

    singer, _ := b.config.SingerGenerator.FromRawPublicKey(memberKey)
    singerHexKey := singer.PublicKeyString()
    if _, voted := b.votedMessage[singerHexKey]; voted {
        return true
    }

    b.votedMessage[singerHexKey] = consensusData
    return true
}

func (b *basicService) verifyNewViewMessage(consensusData SignedConsensusData) (passed bool) {
    passed = false

    justify := consensusData.Justify
    if consensusData.ViewNumber < justify.ViewNumber {
        log.Log.Error("invalid consensus view and justify view.")
        return
    }

    if len(justify.Votes) > 0 {
        passed = b.verifyQCVotes(justify)
    } else {
        passed = true
    }

    return
}


func (b *basicService) verifyPrepareMessage(consensusData SignedConsensusData) (passed bool) {
    passed = false

    //verify view
    justify := consensusData.Justify
    if justify.ViewNumber > consensusData.ViewNumber {
        log.Log.Error("QC view number must small than consensus view number. [ QC@",
            justify.ViewNumber, " : consensus@", consensusData.ViewNumber, " ]")
        return
    }

    //verify vote
    if !b.verifyQCVotes(justify) {
        log.Log.Error("invalid prepare's justify.")
        return
    }

    //verify block
    customerData, err := b.externalProcessor.CustomerDataFromConsensus(consensusData.Payload.CustomerData)
    if err != nil {
        log.Log.Error("get customer data interface from message failed.")
        return
    }

    validCustomerData, err := customerData.Verify()
    if !validCustomerData {
        log.Log.Error("customer data verify failed: ", err)
        return
    }

    validView := true
    if b.lockedQC != nil {
        validView = consensusData.ViewNumber > b.lockedQC.ViewNumber
        if !validView {
            log.Log.Warn("view is not big enough, local: ", b.currentView, " leader: ", consensusData.ViewNumber)
        }
    }

    //check parent hash is extends from highQC first
    parentBytes, _ := structSerializer.ToMFBytes(justify.Payload)
    parentHash := b.config.HashCalc.Sum(parentBytes)
    validExtends := bytes.Equal(parentHash, consensusData.Payload.Parent)
    if !validExtends {
        log.Log.Error("payload is not extends from highQC")
        return
    }

    //check is payload extends from lockedQC
    if b.lockedQC != nil {
        lockedPayloadBytes, _ := structSerializer.ToMFBytes(b.lockedQC.Payload)
        lockedPayloadHash := b.config.HashCalc.Sum(lockedPayloadBytes)
        validExtends = bytes.Equal(lockedPayloadHash, consensusData.Payload.Parent)

        if !validExtends {
            log.Log.Warn("not extends from local locked QC")
        }
    }

    if !validExtends && !validView {
        log.Log.Error("there are no valid views and no valid extended payload")
        return
    }

    if !b.isViewLeader(consensusData.ViewNumber, consensusData.Seal.SignerPublicKey) {
        log.Log.Error("not valid prepare message signature")
        return
    }

    passed = true
    return
}

func (b *basicService) hasEnoughVotes(voteCount int) bool {
    memberCount := len(b.config.Members)
    bftCount := memberCount / 3

    needCount := memberCount - bftCount - 1 //exclude leader self in enough count.

    return needCount <= voteCount
}
