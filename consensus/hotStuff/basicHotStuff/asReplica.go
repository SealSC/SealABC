package basicHotStuff

import (
	"bytes"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
)

func (b *BasicHotStuff) VerifyProposal(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) (passed bool) {
	passed = false

	//verify view
	justify := consensusData.Justify
	if justify.ViewNumber > consensusData.ViewNumber {
		log.Log.Error("QC view number must small than consensus view number. [ QC@",
			justify.ViewNumber, " : consensus@", consensusData.ViewNumber, " ]")
		return
	}

	//verify vote
	if !bs.VerifyQCVotes(justify) {
		log.Log.Error("invalid prepare's justify.")
		return
	}

	//verify block
	customerData, err := bs.ExternalProcessor.CustomerDataFromConsensus(consensusData.Payload.CustomerData)
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
	if bs.LockedQC != nil {
		validView = consensusData.ViewNumber > bs.LockedQC.ViewNumber
		if !validView {
			log.Log.Warn("view is not big enough, local: ", bs.CurrentView, " leader: ", consensusData.ViewNumber)
		}
	}

	//check parent hash is extends from highQC first
	parentBytes, _ := structSerializer.ToMFBytes(justify.Payload)
	parentHash := bs.Config.HashCalc.Sum(parentBytes)
	validExtends := bytes.Equal(parentHash, consensusData.Payload.Parent)
	if !validExtends {
		log.Log.Error("payload is not extends from highQC")
		return
	}

	//check is payload extends from lockedQC
	if bs.LockedQC != nil {
		lockedPayloadBytes, _ := structSerializer.ToMFBytes(bs.LockedQC.Payload)
		lockedPayloadHash := bs.Config.HashCalc.Sum(lockedPayloadBytes)
		validExtends = bytes.Equal(lockedPayloadHash, consensusData.Payload.Parent)
		if !validExtends {
			log.Log.Warn("not extends from local locked QC")
		}
	}

	if !validExtends && !validView {
		log.Log.Error("there are no valid views and no valid extended payload")
		return
	}

	if !bs.IsViewLeader(consensusData.ViewNumber, consensusData.Seal.SignerPublicKey) {
		log.Log.Error("not valid prepare message signature")
		return
	}

	passed = true
	return
}

func (b *BasicHotStuff) OnReceiveProposal(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) {
	voteMsg, err := bs.BuildVoteMessage(consensusData.Phase, consensusData.Payload, consensusData.Id, bs.CurrentView)
	if err != nil {
		log.Log.Error("build vote message failed")
		return
	}

	if bs.CurrentView != consensusData.ViewNumber {
		bs.CurrentView = consensusData.ViewNumber
		log.Log.Warn("local view is not equal network view, but everything on network seems ok, so we sync local view to network view.")
	}

	bs.CurrentPhase = hotStuff.ConsensusPhases.Prepare
	bs.ViewChangeTrigger.Reset(bs.Config.ConsensusTimeout)
	bs.SendMessageToLeader(voteMsg)
}

func (b *BasicHotStuff) BuildNewViewMessage(bs *hotStuff.BasicService, newViewQC hotStuff.QC) (msgPayload []byte, err error) {
	msgPayload, err = bs.BuildConsensusMessage(
		hotStuff.ConsensusPhases.NewView.String(),
		hotStuff.ConsensusPayload{},
		newViewQC,
		"",
		bs.CurrentView)
	return
}

func (b *BasicHotStuff) ProcessCommonPhaseMessage(bs *hotStuff.BasicService, consensusData hotStuff.ConsensusData) {
	allPhases := hotStuff.ConsensusPhases
	switch consensusData.Phase {
	case allPhases.PreCommit.String():
		bs.PrepareQC = &consensusData.Justify
		bs.CurrentPhase = allPhases.PreCommit

	case allPhases.Commit.String():
		bs.LockedQC = &consensusData.Justify
		bs.CurrentPhase = allPhases.Commit

	case allPhases.Decide.String():
		bs.CurrentPhase = allPhases.Decide
	}
}
