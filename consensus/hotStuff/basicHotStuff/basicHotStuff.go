package basicHotStuff

import (
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/log"
)

const MessageFamily = "basic-hot-stuff-consensus"
const MessageVersion = "basic.0.0.1"

type BasicHotStuff struct {
}

func NewBasicHotStuff() (bhs *BasicHotStuff) {
	bhs = &BasicHotStuff{}
	return
}

func (b BasicHotStuff) MessageFamily() string {
	return MessageFamily
}

func (b BasicHotStuff) MessageVersion() string {
	return MessageVersion
}

func (b *BasicHotStuff) RegisterProcessor(bs *hotStuff.BasicService) {
	bs.ConsensusProcessor[hotStuff.MessageTypes.NewView.String()] = bs.GotNewView
	bs.ConsensusProcessor[hotStuff.MessageTypes.Vote.String()] = bs.GotVote
	bs.ConsensusProcessor[hotStuff.MessageTypes.Prepare.String()] = bs.GotPrepare
	bs.ConsensusProcessor[hotStuff.MessageTypes.PreCommit.String()] = bs.GotCommonPhaseMessage
	bs.ConsensusProcessor[hotStuff.MessageTypes.Commit.String()] = bs.GotCommonPhaseMessage
	bs.ConsensusProcessor[hotStuff.MessageTypes.Decide.String()] = bs.GotCommonPhaseMessage
}

func (b *BasicHotStuff) NewRound(bs *hotStuff.BasicService) {
	bs.ViewChangeTrigger.Reset(bs.Config.ConsensusTimeout)
	bs.ClearNewView()

	if !bs.IsCurrentLeader() {
		newViewMsg, err := bs.BuildNewViewMessage()
		if err != nil {
			log.Log.Error("build new view message failed.")
			return
		}

		bs.SendMessageToLeader(newViewMsg)
		return
	} else {
		//log.Log.Println("i am the leader @view ", b.currentView, " use public key: ", b.config.SelfSigner.PublicKeyString())
	}
}
