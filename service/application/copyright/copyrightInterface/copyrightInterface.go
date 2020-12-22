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

package tsInterface

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/dataStructure/enum"
	"SealABC/dataStructure/merkleTree"
	"SealABC/metadata/applicationResult"
	"SealABC/metadata/block"
	"SealABC/metadata/blockchainRequest"
	"SealABC/metadata/seal"
	"SealABC/service"
	"SealABC/service/application/traceableStorage/tsLedger"
	"SealABC/service/system/blockchain/chainStructure"
	"SealABC/storage/db/dbInterface/kvDatabase"
	"SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"encoding/json"
	"sync"
)

var QueryDBType struct{
	KV  enum.Element
	SQL enum.Element
}

type CopyrightStorageApplication struct {
	reqList []string
	reqMap  map[string] blockchainRequest.Entity

	poolLimit int

	poolLock sync.RWMutex
	tsLedger *tsLedger.TSLedger
}

type RequestList struct {
	Requests []blockchainRequest.Entity
}

func (t *CopyrightStorageApplication) Name() (name string) {
	return "Traceable Storage"
}

func (t *CopyrightStorageApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
	reqKey := string(req.Seal.Hash)
	if _, exists := t.reqMap[reqKey]; exists {
		return
	}

	tsReq := tsLedger.TSServiceRequest{}
	err = json.Unmarshal(req.Data, &tsReq)
	if err != nil {
		return
	}

	err = t.tsLedger.VerifyRequest(tsReq)
	if err == nil {
		t.poolLock.RLock()
		t.reqMap[reqKey] = req
		t.reqList = append(t.reqList, reqKey)
		t.poolLock.RUnlock()
	}
	return
}

func (t *CopyrightStorageApplication) Query(reqData []byte) (result interface{}, err error) {
	return t.tsLedger.GetLocalData(string(reqData))
}

func (t *CopyrightStorageApplication) PreExecute(req blockchainRequest.Entity, _ block.Entity) (result []byte, err error) {
	var reqList = RequestList{}
	err = structSerializer.FromMFBytes(req.Data, &reqList)
	if err != nil {
		return
	}

	for _, req := range reqList.Requests {
		_, err = t.PushClientRequest(req)
		if err != nil {
			break
		}
	}
	return
}

func (t *CopyrightStorageApplication) removeTransactionsFromPool(list RequestList) {
	removedReq := map[string] bool {}

	for _, req := range list.Requests {
		reqKey := string(req.Seal.Hash)

		if _, exists := t.reqMap[reqKey]; exists {
			removedReq[reqKey] = true
			delete(t.reqMap, reqKey)
		}
	}

	var newTxPoolRecord []string
	for _, reqHash := range t.reqList {
		if !removedReq[reqHash] {
			newTxPoolRecord = append(newTxPoolRecord, reqHash)
		}
	}

	t.reqList = newTxPoolRecord
}

func (t *CopyrightStorageApplication) Execute (
	req blockchainRequest.Entity,
	blk block.Entity,
	actIndex uint32,
) (result applicationResult.Entity, err error) {
	var reqList = RequestList{}
	err = structSerializer.FromMFBytes(req.Data, &reqList)
	if err != nil {
		return
	}

	for _, req := range reqList.Requests {
		_, err = t.PushClientRequest(req)
		if err != nil {
			break
		}

		tsReq := tsLedger.TSServiceRequest{}
		err = json.Unmarshal(req.Data, &tsReq)
		if err != nil {
			break
		}

		_, err = t.tsLedger.ExecuteRequest(tsReq)
		if err != nil {
			break
		}
	}

	if err == nil {
		t.removeTransactionsFromPool(reqList)
	}
	return
}

func (t *CopyrightStorageApplication) Cancel(req blockchainRequest.Entity) (err error) {
	return
}

func (t *CopyrightStorageApplication) Information() (info service.BasicInformation) {
	info.Name = t.Name()
	info.Description = "this is an traceableStorage application"

	info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
	info.Api.Address = ""
	info.Api.ApiList = []service.ApiInterface {}
	return
}


func (t *CopyrightStorageApplication) RequestsForBlock(_ block.Entity) (reqList []blockchainRequest.Entity, cnt uint32) {
	t.poolLock.Lock()
	defer t.poolLock.Unlock()

	cnt = uint32(len(t.reqList))
	if cnt == 0 {
		return
	}

	tsReqList := RequestList{}
	mt := merkleTree.Tree{}
	for idx, reqKey := range t.reqList {
		req := t.reqMap[reqKey]
		mt.AddHash(req.Seal.Hash)
		tsReqList.Requests = append(tsReqList.Requests, req)

		if idx >= (t.poolLimit - 1) {
			cnt = uint32(idx)
			break
		}
	}

	txRoot, _ := mt.Calculate()

	blkReqData, _ := structSerializer.ToMFBytes(tsReqList)
	packedReq := blockchainRequest.Entity{
		EntityData: blockchainRequest.EntityData{
			RequestApplication: t.Name(),
			RequestAction:      "",
			Data:               blkReqData,
			QueryString:        "",
		},

		Packed:      true,
		PackedCount: cnt,
		Seal:       seal.Entity{
			Hash:            txRoot, //use merkle tree root as seal hash for packed actions
			Signature:       nil,
			SignerPublicKey: nil,
			SignerAlgorithm: "",
		},
	}

	return []blockchainRequest.Entity{packedReq}, 1
}


func (t *CopyrightStorageApplication) SetBlockchainService(_ interface{}) {}
func (t *CopyrightStorageApplication) UnpackingActionsAsRequests(_ blockchainRequest.Entity) ([]blockchainRequest.Entity, error) {return nil, nil}

func (t *CopyrightStorageApplication) GetActionAsRequest(req blockchainRequest.Entity) (newReq blockchainRequest.Entity) {
	return 
}

func Load()  {
	tsLedger.Load()
}

func NewApplicationInterface(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver) (app chainStructure.IBlockchainExternalApplication) {
	ts := CopyrightStorageApplication{
		reqList:   []string {},
		reqMap:    map[string]blockchainRequest.Entity{},
		poolLock:  sync.RWMutex{},
		poolLimit: 1000,
		tsLedger:  tsLedger.NewTraceableStorage(kvDriver, sqlDriver),
	}

	app = &ts
	return
}

