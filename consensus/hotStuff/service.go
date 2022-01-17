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
	"github.com/SealSC/SealABC/consensus"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/message"
	"github.com/SealSC/SealABC/network"
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type consensusProcessor func(consensusData SignedConsensusData) (reply *message.Message)

type basicHotStuffInformation struct {
	Network           network.StaticInformation
	Members           []string
	ConsensusInterval time.Duration
	ConsensusTimeout  time.Duration
}

type basicService struct {
	config Config

	currentState enum.Element
	currentPhase enum.Element
	phaseLock    sync.Mutex

	newViews          map[string]SignedConsensusData
	votedMessage      map[string]SignedConsensusData
	prepareQC         *QC
	lockedQC          *QC
	viewChangeTrigger *time.Timer
	currentView       uint64

	consensusProcessor map[string]consensusProcessor
	externalProcessor  consensus.ExternalProcessor

	network network.IService

	information *basicHotStuffInformation
}

var Basic basicService

func (b *basicService) isCurrentLeader() (isLeader bool) {
	selfKey := b.config.SelfSigner.PublicKeyBytes()
	leader := b.getLeader()

	return bytes.Equal(selfKey, leader.Signer.PublicKeyBytes())
}

func (b *basicService) isViewLeader(viewNumber uint64, key []byte) (isLeader bool) {
	leaderIndex := (viewNumber + 1) % uint64(len(b.config.Members))
	leader := b.config.Members[leaderIndex]
	return bytes.Equal(key, leader.Signer.PublicKeyBytes())
}

func (b *basicService) getLeader() (leader Member) {
	leaderIndex := (b.currentView + 1) % uint64(len(b.config.Members))
	leader = b.config.Members[leaderIndex]
	return
}

func (b *basicService) clearNewView() {
	var keyForDel []string
	for k, v := range b.newViews {
		if b.currentView != v.ViewNumber {
			keyForDel = append(keyForDel, k)
		}
	}

	for _, k := range keyForDel {
		delete(b.newViews, k)
	}
}

func (b *basicService) newRound() {
	b.viewChangeTrigger.Reset(b.config.ConsensusTimeout)
	b.clearNewView()

	if !b.isCurrentLeader() {
		newViewMsg, err := b.buildNewViewMessage()
		if err != nil {
			log.Log.Error("build new view message failed.")
			return
		}
		go b.sendMessageToLeader(newViewMsg)
		return
	} else {
		//log.Log.Println("i am the leader @view ", b.currentView, " use public key: ", b.config.SelfSigner.PublicKeyString())
	}
}

func (b *basicService) startViewChangeMonitor() {
	log.Log.Println("start view change monitor : ", b.config.ConsensusTimeout)
	b.currentState = consensus.States.Running
	b.viewChangeTrigger = time.NewTimer(b.config.ConsensusTimeout)

	for {
		select {
		case <-b.viewChangeTrigger.C:
			//do view change
			b.viewChange()

			//reset the timer
			b.viewChangeTrigger.Reset(b.config.ConsensusTimeout)
		}
	}
}

func (b *basicService) viewChange() {
	b.phaseLock.Lock()
	defer b.phaseLock.Unlock()

	b.currentView += 1
	b.currentPhase = consensusPhases.NewView
	log.Log.Println("view change to new view ", b.currentView)

	b.newRound()
}

func (b *basicService) initService() {
	onlineCheck := time.NewTimer(b.config.MemberOnlineCheckInterval)
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

		onlineCheck.Reset(b.config.MemberOnlineCheckInterval)
	}

	if b.currentState.String() == consensus.States.Stopped.String() {
		return
	}

	if !allMemberOnline {
		return
	}

	b.phaseLock.Lock()
	defer b.phaseLock.Unlock()
	time.Sleep(time.Millisecond * 100)
	go b.startViewChangeMonitor()

	b.newRound()
}

func (b *basicService) Feed(msg message.Message) (reply *message.Message) {
	if msg.Family != MessageFamily {
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

	b.phaseLock.Lock()
	defer b.phaseLock.Unlock()

	//todo: modular log system
	//log.Log.Println("got message: ", msg.Type)
	if handle, exists := b.consensusProcessor[msg.Type]; exists {
		reply = handle(consensusData)
	}
	//log.Log.Println("message handle over ")

	return
}

func (b *basicService) Start(cfg interface{}) (err error) {
	config, ok := cfg.(Config)
	if !ok {
		err = errors.New("invalid config")
		return
	}

	Basic.config = config
	Basic.currentState = consensus.States.Init

	go b.initService()
	return
}

func (b *basicService) Stop() (err error) {
	return
}

func (b *basicService) RegisterExternalProcessor(processor consensus.ExternalProcessor) {
	b.externalProcessor = processor
	return
}

func (b *basicService) GetMessageFamily() (family string) {
	family = MessageFamily
	return
}

func (b *basicService) GetExternalProcessor() (processor consensus.ExternalProcessor) {
	processor = b.externalProcessor
	return
}

func (b *basicService) GetConsensusCustomerData(msg message.Message) (data []byte, err error) {
	consensusData := SignedConsensusData{}
	err = json.Unmarshal(msg.Payload, &consensusData)
	if err != nil {
		return
	}
	data = consensusData.ConsensusData.Justify.Payload.CustomerData
	return
}

func (b *basicService) Load(networkService network.IService, processor consensus.ExternalProcessor) {
	enum.Build(&messageTypes, 0, "basic-hot-stuff-")
	enum.Build(&consensusPhases, 0, "")

	b.viewChangeTrigger = time.NewTimer(b.config.ConsensusTimeout)

	b.network = networkService

	b.newViews = map[string]SignedConsensusData{}
	b.votedMessage = map[string]SignedConsensusData{}
	b.consensusProcessor = map[string]consensusProcessor{}

	b.externalProcessor = processor

	b.registerLeaderProcessor()
	b.registerReplicaProcessor()
}

func (b *basicService) StaticInformation() interface{} {
	if b.information != nil {
		return b.information
	}

	info := basicHotStuffInformation{}

	info.Network = b.network.StaticInformation()
	info.Members = b.allMembersKey()
	info.ConsensusInterval = b.config.ConsensusInterval
	info.ConsensusTimeout = b.config.ConsensusTimeout

	if !b.isAllMembersOnline() {
		return info
	}

	b.information = &info
	return b.information
}
