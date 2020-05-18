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

package engineService

import (
    "SealABC/consensus"
    "SealABC/dataStructure/enum"
    "SealABC/log"
    "SealABC/metadata/serviceRequest"
    "encoding/json"
)

type consensusProcessor struct {}

type requestList struct {
    Requests [][]byte
}

func (r *requestList) Verify() (passed bool, err error) {
    passed = true

    for _, req := range r.Requests {
        srvReq := serviceRequest.Entity{}
        err = json.Unmarshal(req, &srvReq)
        if err != nil {
            continue
        }
        _, err = preExecuteServiceRequest(srvReq)
        //todo: handle the result of pre-execute

        if err != nil {
            passed = false
            return
        }
    }
    return
}

func (r requestList) Bytes() (data []byte, err error) {
    data, err = json.Marshal(r)
    return
}

func (e *consensusProcessor) EventProcessor(event enum.Element, customerData []byte) {
    if event.String() == consensus.Event.Failed.String() {
        //todo: handle consensus failed
        log.Log.Error("consensus failed")
        return
    }

    reqList := requestList{}

    err := json.Unmarshal(customerData, &reqList)
    if err != nil {
        log.Log.Error("deserialize consensus data failed: ", customerData)
        return
    }

    for _, req := range reqList.Requests {
        srvReq := serviceRequest.Entity{}
        err = json.Unmarshal(req, &srvReq)
        if err != nil {
            log.Log.Error("deserialize consensus data failed: ", err.Error())
            continue
        }


        _, err = executeServiceRequest(srvReq)
        //todo: handle the result of pre-execute

        if err != nil {
            log.Log.Error("execute request failed.")
            continue
        }
        //todo: need process result of the action
        //log.Log.Println("need process result of the action when decide: ", result.ResultData)
    }
    return
}

func (e *consensusProcessor) CustomerDataToConsensus() (data consensus.ICustomerData, _ error) {
    allValidRequests := getAllRequestsNeedConsensus()
    data = &requestList{
        allValidRequests,
    }
    return
}

func (e *consensusProcessor) CustomerDataFromConsensus(data []byte) (customerData consensus.ICustomerData, err error) {
    reqList := requestList{}

    err = json.Unmarshal(data, &reqList)
    if err != nil {
        return
    }

    customerData = &reqList
    return
}

func (e *consensusProcessor) NewDataBasedOnConsensus(data consensus.ICustomerData) (newData consensus.ICustomerData, err error) {
    //todo: prepare for parallel service
    return
}

var ConsensusProcessor consensusProcessor

//func Start(config engineStartup.EngineConfig) {
//    engineStartup.EngineStart(config, &instance)
//    return
//}
