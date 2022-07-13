package consensusFactory

import (
	"github.com/SealSC/SealABC/consensus"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/consensus/hotStuff/basicHotStuff"
	"github.com/SealSC/SealABC/consensus/hotStuff/chainedHotStuff"
	"github.com/SealSC/SealABC/log"
)

func NewConsensusService(t consensus.Type) (service consensus.IConsensusService) {
	consensusType := consensus.BasicHotStuff
	if t != "" {
		consensusType = t
	}

	switch consensusType {
	case consensus.BasicHotStuff:
		service = hotStuff.NewHotStuff(basicHotStuff.NewBasicHotStuff())
	case consensus.ChainedHotStuff:
		service = hotStuff.NewHotStuff(chainedHotStuff.NewChainedHotStuff())
	default:
		service = hotStuff.NewHotStuff(basicHotStuff.NewBasicHotStuff())
	}

	log.Log.Infof("ConsensusType:%s", consensusType)
	return
}
