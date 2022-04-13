package chainedHotStuff

import (
	"bytes"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
)

func (c *ChainedHotStuff) VerifyProposal(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) (passed bool) {
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
func (c *ChainedHotStuff) OnProposal(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) {
	bs.CurrentPhase = hotStuff.ConsensusPhases.Generic
	c.update(bs, consensusData.ConsensusData)
}

func (c *ChainedHotStuff) BuildNewViewMessage(bs *hotStuff.BasicService, newViewQC hotStuff.QC) (msgPayload []byte, err error) {
	parentId := ""
	if bs.PrepareQC != nil {
		parentId = bs.PrepareQC.NodeId
	}
	msgPayload, err = bs.BuildConsensusMessage(
		hotStuff.ConsensusPhases.NewView.String(),
		hotStuff.ConsensusPayload{},
		newViewQC,
		parentId,
		bs.CurrentView)
	return
}

func (c *ChainedHotStuff) GotVoteRule(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) bool {
	return bs.IsViewLeader(consensusData.ViewNumber, bs.Config.SelfSigner.PublicKeyBytes())
}

func (c *ChainedHotStuff) GotVote(bs *hotStuff.BasicService, consensusData hotStuff.ConsensusData) {
	defer func() {
		bs.VotedMessage = map[string]hotStuff.SignedConsensusData{}
	}()

	votedQC := hotStuff.QC{}
	votedQC.Payload = consensusData.Payload

	votedQC.ViewNumber = bs.CurrentView
	votedQC.Phase = consensusData.Phase
	votedQC.NodeId = consensusData.ParentId

	for _, votedMsg := range bs.VotedMessage {
		votedQC.Votes = append(votedQC.Votes, votedMsg.Seal)
	}

	selfVote, err := bs.BuildVote(votedQC.QCData)
	if err != nil {
		return
	}
	votedQC.Votes = append(votedQC.Votes, selfVote)

	err = bs.BroadCastPrepareMessage(votedQC)
	if err != nil {
		return
	}
}
