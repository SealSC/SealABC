package chainedHotStuff

import (
	"github.com/SealSC/SealABC/consensus"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
	"time"
)

const MessageFamily = "chained-hot-stuff-consensus"
const MessageVersion = "chained.0.0.1"

type ChainedHotStuff struct {
	nodeMap   map[string]hotStuff.ConsensusData
	pacemaker Pacemaker
}

func NewChainedHotStuff() (chs *ChainedHotStuff) {
	chs = &ChainedHotStuff{}
	chs.pacemaker = &RoundRobinPM{}
	chs.nodeMap = map[string]hotStuff.ConsensusData{}

	return
}

func (c *ChainedHotStuff) MessageFamily() string {
	return MessageFamily
}

func (c *ChainedHotStuff) MessageVersion() string {
	return MessageVersion
}

func (c *ChainedHotStuff) RegisterProcessor(bs *hotStuff.BasicService) {
	bs.ConsensusProcessor[hotStuff.MessageTypes.NewView.String()] = bs.GotNewView
	bs.ConsensusProcessor[hotStuff.MessageTypes.Vote.String()] = bs.GotVote
	bs.ConsensusProcessor[hotStuff.MessageTypes.Generic.String()] = bs.GotGeneric
}

func (c *ChainedHotStuff) update(bs *hotStuff.BasicService, node hotStuff.ConsensusData) {
	defer func() {
		c.advanceView(bs)
	}()

	c.saveNode(node)
	bs.UpdateBLeaf(node)

	if !bs.IsNextViewLeader(node.ViewNumber, bs.Config.SelfSigner.PublicKeyBytes()) {
		preView := bs.CurrentView
		go func() {
			time.Sleep(bs.Config.ConsensusInterval)
			bs.PhaseLock.Lock()
			defer bs.PhaseLock.Unlock()

			if preView != bs.CurrentView-1 {
				log.Log.Infof("outdated message, viewNumber: %d", preView)
				return
			}

			c.sendVoteToNextLeader(bs, node, bs.CurrentView)
		}()
	}

	var nodeId string
	var prepare, preCommit, commit, decide hotStuff.ConsensusData

	prepare = node
	exist := false
	nodeId = prepare.Justify.NodeId

	preCommit, exist = c.getNode(nodeId)

	if exist {
		nodeId = preCommit.Justify.NodeId
		commit, exist = c.getNode(nodeId)
	}

	if exist {
		nodeId = commit.Justify.NodeId
		decide, exist = c.getNode(nodeId)
	}

	if !c.commitRule(prepare, preCommit) {
		return
	}
	c.pacemaker.UpdateHighQC(bs, prepare)

	if !c.commitRule(preCommit, commit) {
		return
	}
	bs.LockedQC = &preCommit.Justify

	if !c.commitRule(commit, decide) {
		return
	}

	c.PruneToHeight(decide.ViewNumber)

	if bs.ExternalProcessor != nil {
		bs.ExternalProcessor.EventProcessor(consensus.Event.Success, decide.Payload.CustomerData)
	}
}

func (c *ChainedHotStuff) advanceView(bs *hotStuff.BasicService) {
	c.pacemaker.AdvanceView(bs)
}

func (c *ChainedHotStuff) NewRound(bs *hotStuff.BasicService) {
	c.pacemaker.OnNextSyncView(bs)
}

func (c *ChainedHotStuff) sendVoteToNextLeader(bs *hotStuff.BasicService, node hotStuff.ConsensusData, viewNumber uint64) {
	voteMsg, err := bs.BuildVoteMessage(node.Phase, node.Payload, node.Id, viewNumber)
	if err != nil {
		log.Log.Error("build vote message failed")
		return
	}

	bs.SendMessageToLeader(voteMsg)
}

func (c *ChainedHotStuff) PruneToHeight(height uint64) {
	var keyForDel []string
	for k, v := range c.nodeMap {
		if v.ViewNumber < height {
			keyForDel = append(keyForDel, k)
		}
	}

	for _, k := range keyForDel {
		delete(c.nodeMap, k)
	}
}

func (c *ChainedHotStuff) saveNode(consensusData hotStuff.ConsensusData) {
	c.nodeMap[consensusData.Id] = consensusData
}

func (c *ChainedHotStuff) getNode(id string) (node hotStuff.ConsensusData, exist bool) {
	node, exist = c.nodeMap[id]
	return
}

func (c *ChainedHotStuff) commitRule(child hotStuff.ConsensusData, parent hotStuff.ConsensusData) (passed bool) {
	passed = false
	if !child.IsStructureEmpty() && !parent.IsStructureEmpty() && child.ParentId == parent.Id {
		passed = true
	}
	return
}
