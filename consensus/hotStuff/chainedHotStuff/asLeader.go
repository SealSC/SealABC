package chainedHotStuff

import (
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
)

func (c *ChainedHotStuff) NewView(bs *hotStuff.BasicService, consensusData hotStuff.SignedConsensusData) {
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

func (c *ChainedHotStuff) Proposal(bs *hotStuff.BasicService, highQC hotStuff.QC, node hotStuff.ConsensusPayload) (err error) {

	parentId := highQC.NodeId
	parentNode, exist := c.getNode(parentId)
	if exist {
		tempView := parentNode.ViewNumber + 1
		for tempView < bs.CurrentView {
			dummyNode := hotStuff.ConsensusData{}
			dummyNode.ViewNumber = tempView
			dummyNode.Justify = hotStuff.QC{}
			dummyNode.Id = dummyNode.NodeId()
			dummyNode.ParentId = parentId
			c.saveNode(dummyNode)
			parentId = dummyNode.Id
			tempView++
		}
	}

	n := hotStuff.ConsensusData{}
	n.ViewNumber = bs.CurrentView
	n.Phase = hotStuff.ConsensusPhases.Generic.String()
	n.Justify = highQC
	n.Payload = node
	n.ParentId = parentId
	n.Id = n.NodeId()

	msgPayload, err := bs.BuildConsensusMessage(
		n.Phase,
		n.Payload,
		n.Justify,
		n.ParentId,
		bs.CurrentView)

	if err != nil {
		return
	}
	msg := bs.BuildCommonMessage(hotStuff.MessageTypes.Generic, msgPayload)
	bs.BroadCastMessage(msg)
	c.update(bs, n)

	return
}
