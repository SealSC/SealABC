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

package consensus

import (
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/message"
	"github.com/SealSC/SealABC/network"
)

type state struct {
	Init    enum.Element
	Running enum.Element
	Stopped enum.Element
}

type event struct {
	Success enum.Element
	Failed  enum.Element
}

type ICustomerData interface {
	Verify([]byte) (passed bool, err error)
	Bytes() ([]byte, error)
}

var States state
var Event event

type ExternalProcessor interface {
	EventProcessor(event enum.Element, customerData []byte)
	NewDataBasedOnConsensus(data ICustomerData) (newData ICustomerData, err error)
	CustomerDataToConsensus(lastCustomerData []byte) (data ICustomerData, err error)
	CustomerDataFromConsensus(data []byte) (customData ICustomerData, err error)
}

type IConsensusService interface {
	Load(networkService network.IService, processor ExternalProcessor)
	Start(cfg interface{}) (err error)
	Stop() (err error)

	Feed(msg message.Message) (reply *message.Message)
	RegisterExternalProcessor(processor ExternalProcessor)

	GetMessageFamily() string
	GetExternalProcessor() (processor ExternalProcessor)
	GetConsensusCustomerData(msg message.Message) (data []byte, err error)
	GetLastConsensusCustomerData() []byte
	StaticInformation() interface{}
}

func Load(service IConsensusService, ns network.IService, processor ExternalProcessor) IConsensusService {
	enum.Build(&States, 0, "")
	enum.Build(&Event, 0, "")
	service.Load(ns, processor)
	Driver.consensusRegister(service, ns)

	return service
}
