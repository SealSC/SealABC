package chainedHotStuff

import (
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
)

type Pacemaker interface {
	AdvanceView(*hotStuff.BasicService)
	UpdateHighQC(*hotStuff.BasicService, hotStuff.ConsensusData)
	OnNextSyncView(*hotStuff.BasicService)
	OnReceiveNewView(*hotStuff.BasicService, hotStuff.SignedConsensusData)
}

type RoundRobinPM struct {
}

func (r *RoundRobinPM) AdvanceView(bs *hotStuff.BasicService) {
	bs.CurrentView += 1
	bs.ViewChangeTrigger.Reset(bs.Config.ConsensusTimeout)
}

func (r *RoundRobinPM) UpdateHighQC(bs *hotStuff.BasicService, consensusData hotStuff.ConsensusData) {
	bs.PrepareQC = &consensusData.Justify
}

func (r *RoundRobinPM) OnNextSyncView(bs *hotStuff.BasicService) {
	bs.ViewChangeTrigger.Reset(bs.Config.ConsensusTimeout)
	bs.ClearNewView()
	bs.ClearPrepare()
	bs.ClearBLeaf()

	if !bs.IsCurrentLeader() {
		newViewMsg, err := bs.BuildNewViewMessage()
		if err != nil {
			log.Log.Error("build new view message failed.")
			return
		}

		go bs.SendMessageToLeader(newViewMsg)
		return
	}
}

func (r *RoundRobinPM) OnReceiveNewView(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) {
	singer, _ := bs.Config.SingerGenerator.FromRawPublicKey(consensusData.Seal.SignerPublicKey)

	bs.NewViews[singer.PublicKeyString()] = consensusData

	if !bs.HasEnoughVotes(len(bs.NewViews)) {
		return
	}

	bs.CurrentView = consensusData.ViewNumber
	highQC := bs.PickHighQC()

	highQC.NodeId = consensusData.ParentId
	bs.PrepareQC = &highQC

	err := bs.BroadCastPrepareMessage(highQC)

	if err != nil {
		log.Log.Error("build prepare message failed.")
		bs.NewViews = map[string]hotStuff.SignedConsensusData{}
		return
	}

	bs.NewViews = map[string]hotStuff.SignedConsensusData{}

	return
}
