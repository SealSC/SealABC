package basicHotStuff

import (
	"github.com/SealSC/SealABC/consensus"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
)

func (b *BasicHotStuff) NewView(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) {

	singer, _ := bs.Config.SingerGenerator.FromRawPublicKey(consensusData.Seal.SignerPublicKey)

	bs.NewViews[singer.PublicKeyString()] = consensusData

	if !bs.HasEnoughVotes(len(bs.NewViews)) {
		return
	}

	//if consensusData.ViewNumber == b.currentView && b.currentPhase != consensusPhases.NewView {
	//    return
	//}

	bs.CurrentView = consensusData.ViewNumber
	highQC := bs.PickHighQC()

	err := bs.BroadCastPrepareMessage(highQC)

	if err != nil {
		log.Log.Error("build prepare message failed.")
		bs.NewViews = map[string]hotStuff.SignedConsensusData{}
		return
	}

	bs.NewViews = map[string]hotStuff.SignedConsensusData{}
	bs.CurrentPhase = hotStuff.ConsensusPhases.Prepare

	return
}

func (b *BasicHotStuff) GotVoteRule(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) bool {
	return true
}

func (b *BasicHotStuff) GotVote(bs *hotStuff.BasicService, consensusData hotStuff.ConsensusData) {
	nextPhase, msgType := bs.GetNextPhaseAndMsgType()

	//todo: pre-commit message need rebuild payload from all votes for parallel service
	votedQC := hotStuff.QC{}
	votedQC.Payload = consensusData.Payload
	votedQC.ViewNumber = bs.CurrentView
	votedQC.Phase = consensusData.Phase
	votedQC.NodeId = consensusData.Id

	nextPhaseMsg, err := bs.BuildCommonPhaseMessage(nextPhase, msgType, &votedQC)
	if err != nil {
		return
	}

	switch bs.CurrentPhase {
	case hotStuff.ConsensusPhases.PreCommit:
		bs.PrepareQC = &votedQC

	case hotStuff.ConsensusPhases.Commit:
		bs.LockedQC = &votedQC
	}

	bs.CurrentPhase = nextPhase
	bs.VotedMessage = map[string]hotStuff.SignedConsensusData{}
	bs.BroadCastMessage(nextPhaseMsg)

	if nextPhase == hotStuff.ConsensusPhases.Decide {
		if bs.ExternalProcessor != nil {
			bs.ExternalProcessor.EventProcessor(consensus.Event.Success, consensusData.Payload.CustomerData)
		}

		bs.CurrentPhase = hotStuff.ConsensusPhases.NewView
		bs.CurrentView += 1

		bs.NewRound()
	}
	return
}

func (b *BasicHotStuff) Proposal(bs *hotStuff.BasicService, highQC hotStuff.QC, node hotStuff.ConsensusPayload) (err error) {
	msgPayload, err := bs.BuildConsensusMessage(
		hotStuff.ConsensusPhases.Prepare.String(),
		node,
		highQC,
		highQC.NodeId,
		bs.CurrentView)

	if err != nil {
		return
	}

	msg := bs.BuildCommonMessage(hotStuff.MessageTypes.Prepare, msgPayload)

	bs.BroadCastMessage(msg)
	return
}
