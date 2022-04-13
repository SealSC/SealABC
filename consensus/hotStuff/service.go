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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SealSC/SealABC/consensus"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/message"
	"github.com/SealSC/SealABC/network"
	"sync"
	"time"
)

type consensusProcessor func(consensusData SignedConsensusData) (reply *message.Message)

type hotStuff interface {
	MessageFamily() string
	MessageVersion() string
	RegisterProcessor(*BasicService)
	NewRound(*BasicService)
	NewView(*BasicService, SignedConsensusData)
	Proposal(*BasicService, QC, ConsensusPayload) (err error)
	VerifyProposal(*BasicService, SignedConsensusData) (passed bool)
	OnProposal(*BasicService, SignedConsensusData)
	GotVoteRule(*BasicService, SignedConsensusData) bool
	GotVote(*BasicService, ConsensusData)
	BuildNewViewMessage(*BasicService, QC) (msgPayload []byte, err error)
	GetLastProposal() []byte
}

type basicHotStuffInformation struct {
	Network           network.StaticInformation
	Members           []string
	ConsensusInterval time.Duration
	ConsensusTimeout  time.Duration
}

type BasicService struct {
	Config Config

	currentState enum.Element
	CurrentPhase enum.Element
	PhaseLock    sync.Mutex

	NewViews     map[string]SignedConsensusData
	VotedMessage map[string]SignedConsensusData

	PrepareQC         *QC //GenericQC for Chained
	LockedQC          *QC
	ViewChangeTrigger *time.Timer
	CurrentView       uint64

	ConsensusProcessor map[string]consensusProcessor
	ExternalProcessor  consensus.ExternalProcessor

	network network.IService

	information *basicHotStuffInformation
	hotStuff    hotStuff
}

var Basic BasicService

func NewHotStuff(hotStuff hotStuff) *BasicService {
	Basic.hotStuff = hotStuff
	return &Basic
}

func (b *BasicService) IsCurrentLeader() (isLeader bool) {
	selfKey := b.Config.SelfSigner.PublicKeyBytes()
	leader := b.getLeader()

	return bytes.Equal(selfKey, leader.Signer.PublicKeyBytes())
}

func (b *BasicService) IsViewLeader(viewNumber uint64, key []byte) (isLeader bool) {
	leaderIndex := (viewNumber + 1) % uint64(len(b.Config.Members))
	leader := b.Config.Members[leaderIndex]
	return bytes.Equal(key, leader.Signer.PublicKeyBytes())
}

func (b *BasicService) IsNextViewLeader(viewNumber uint64, key []byte) (isLeader bool) {
	leaderIndex := (viewNumber + 2) % uint64(len(b.Config.Members))
	leader := b.Config.Members[leaderIndex]
	return bytes.Equal(key, leader.Signer.PublicKeyBytes())
}

func (b *BasicService) getLeader() (leader Member) {
	leaderIndex := (b.CurrentView + 1) % uint64(len(b.Config.Members))
	leader = b.Config.Members[leaderIndex]
	return
}

func (b *BasicService) ClearNewView() {
	var keyForDel []string
	for k, v := range b.NewViews {
		if b.CurrentView != v.ViewNumber {
			keyForDel = append(keyForDel, k)
		}
	}

	for _, k := range keyForDel {
		delete(b.NewViews, k)
	}
}
func (b *BasicService) ClearPrepare() {
	b.PrepareQC = nil
}

func (b *BasicService) NewRound() {
	b.hotStuff.NewRound(b)
}

func (b *BasicService) startViewChangeMonitor() {
	log.Log.Println("start view change monitor : ", b.Config.ConsensusTimeout)
	b.currentState = consensus.States.Running
	b.ViewChangeTrigger = time.NewTimer(b.Config.ConsensusTimeout)

	for {
		select {
		case <-b.ViewChangeTrigger.C:
			//do view change
			b.viewChange()

			//reset the timer
			b.ViewChangeTrigger.Reset(b.Config.ConsensusTimeout)
		}
	}
}

func (b *BasicService) viewChange() {
	b.PhaseLock.Lock()
	defer b.PhaseLock.Unlock()

	b.CurrentView += 1
	b.CurrentPhase = ConsensusPhases.NewView
	log.Log.Println("view change to new view ", b.CurrentView)

	b.NewRound()
}

func (b *BasicService) initService() {
	onlineCheck := time.NewTimer(b.Config.MemberOnlineCheckInterval)
	allMemberOnline := false
	for {
		if b.currentState.String() == consensus.States.Stopped.String() {
			break
		}

		select {
		case <-onlineCheck.C:
			allMemberOnline = b.isAllMembersOnline()
		}

		if allMemberOnline {
			log.Log.Println("all members online now!")
			break
		}

		onlineCheck.Reset(b.Config.MemberOnlineCheckInterval)
	}

	if b.currentState.String() == consensus.States.Stopped.String() {
		return
	}

	if !allMemberOnline {
		return
	}

	b.PhaseLock.Lock()
	defer b.PhaseLock.Unlock()
	time.Sleep(time.Millisecond * 300)
	go b.startViewChangeMonitor()

	b.NewRound()
}

func (b *BasicService) Feed(msg message.Message) (reply *message.Message) {
	if msg.Family != b.hotStuff.MessageFamily() {
		return
	}

	consensusData, err := b.consensusDataFromMessage(msg)
	if err != nil {
		return
	}

	if !b.isMemberKey(consensusData.Seal.SignerPublicKey) {
		log.Log.Error("not a member of this consensus network")
		return
	}

	dataForSign := QCData{
		Phase:      consensusData.Phase,
		ViewNumber: consensusData.ViewNumber,
		Payload:    consensusData.Payload,
	}

	//todo: will be verify hash, not the data directly
	if !b.verifySignature(dataForSign, consensusData.Seal) {
		log.Log.Error("invalid vote message signature")
		return
	}

	b.PhaseLock.Lock()
	defer b.PhaseLock.Unlock()

	//todo: modular log system
	log.Log.Println("got message: ", msg.Type)
	if handle, exists := b.ConsensusProcessor[msg.Type]; exists {
		reply = handle(consensusData)
	}
	//log.Log.Println("message handle over ")

	return
}

func (b *BasicService) Start(cfg interface{}) (err error) {
	config, ok := cfg.(Config)
	if !ok {
		errors.New("invalid config")
		return
	}

	Basic.Config = config
	Basic.currentState = consensus.States.Init

	go b.initService()
	return
}

func (b *BasicService) Stop() (err error) {
	return
}

func (b *BasicService) RegisterExternalProcessor(processor consensus.ExternalProcessor) {
	b.ExternalProcessor = processor
	return
}

func (b *BasicService) GetMessageFamily() (family string) {
	family = b.hotStuff.MessageFamily()
	return
}

func (b *BasicService) GetExternalProcessor() (processor consensus.ExternalProcessor) {
	processor = b.ExternalProcessor
	return
}

func (b *BasicService) GetConsensusCustomerData(msg message.Message) (data []byte, err error) {
	consensusData := SignedConsensusData{}
	err = json.Unmarshal(msg.Payload, &consensusData)
	if err != nil {
		return
	}
	data = consensusData.ConsensusData.Justify.Payload.CustomerData
	return
}

func (b *BasicService) GetLastConsensusCustomerData() []byte {
	return b.hotStuff.GetLastProposal()
}

func (b *BasicService) Load(networkService network.IService, processor consensus.ExternalProcessor) {
	enum.Build(&MessageTypes, 0, fmt.Sprintf("%s-", b.hotStuff.MessageFamily()))
	enum.Build(&ConsensusPhases, 0, "")

	b.ViewChangeTrigger = time.NewTimer(b.Config.ConsensusTimeout)

	b.network = networkService

	b.NewViews = map[string]SignedConsensusData{}
	b.VotedMessage = map[string]SignedConsensusData{}
	b.ConsensusProcessor = map[string]consensusProcessor{}

	b.ExternalProcessor = processor

	b.hotStuff.RegisterProcessor(b)
}

func (b *BasicService) StaticInformation() interface{} {
	if b.information != nil {
		return b.information
	}

	info := basicHotStuffInformation{}

	info.Network = b.network.StaticInformation()
	info.Members = b.allMembersKey()
	info.ConsensusInterval = b.Config.ConsensusInterval
	info.ConsensusTimeout = b.Config.ConsensusTimeout

	if !b.isAllMembersOnline() {
		return info
	}

	b.information = &info
	return b.information
}
