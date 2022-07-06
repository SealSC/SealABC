package consensusFactory

import (
	"github.com/SealSC/SealABC/consensus"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/consensus/hotStuff/basicHotStuff"
	"github.com/SealSC/SealABC/consensus/hotStuff/chainedHotStuff"
	"github.com/SealSC/SealABC/log"
)

func NewConsensusService(consensusType consensus.Type) (service consensus.IConsensusService) {
	if consensusType == "" {
		consensusType = consensus.BasicHotStuff
	}

	switch consensusType {
	case consensus.BasicHotStuff:
		service = hotStuff.NewHotStuff(basicHotStuff.NewBasicHotStuff())
	case consensus.ChainedHotStuff:
		service = hotStuff.NewHotStuff(chainedHotStuff.NewChainedHotStuff())
	}

	log.Log.Infof("ConsensusType:%s", consensusType)
	return
}
