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

package uidInterface

import (
	"errors"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/dataStructure/merkleTree"
	"github.com/SealSC/SealABC/metadata/applicationResult"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/metadata/seal"
	"github.com/SealSC/SealABC/service"
	"github.com/SealSC/SealABC/service/application/universalIdentification/uidLedger"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"sync"
)

type ActionList struct {
	Actions []blockchainRequest.Entity
}

type UniversalIdentificationApplication struct {
	reqPool map[string] blockchainRequest.Entity
	reqList [] blockchainRequest.Entity

	poolLock sync.Mutex

	ledger uidLedger.UIDLedger
}

func (u *UniversalIdentificationApplication) Name() (name string) {
	return "Universal Identification"
}

func (u *UniversalIdentificationApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
	result, err = u.ledger.VerifyAction(req.RequestAction, req.Data)

	if err != nil {
		return
	}

	//push request to pool
	u.poolLock.Lock()
	u.reqPool[string(req.Seal.Hash)] = req
	u.reqList = append(u.reqList, req)
	u.poolLock.Unlock()

	return
}

func (u *UniversalIdentificationApplication) Query(req []byte) (result interface{}, err error) {
	return
}

func (u *UniversalIdentificationApplication) PreExecute(req blockchainRequest.Entity, _ block.Entity) (result []byte, err error) {
	reqList := ActionList{}
	err = structSerializer.FromMFBytes(req.Data, &reqList)
	if err != nil {
		return
	}

	u.poolLock.Lock()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}

		u.poolLock.Unlock()
	}()

	for _, req := range reqList.Actions {
		reqKey := string(req.Seal.Hash)
		if _, exist := u.reqPool[reqKey]; exist {
			continue
		}

		_, err = u.ledger.VerifyAction(req.RequestAction, req.Data)
		if err != nil {
			return
		}

		u.reqPool[reqKey] = req
		u.reqList = append(u.reqList, req)
	}

	return
}

func (u *UniversalIdentificationApplication) removeRequestFromPool(reqList []blockchainRequest.Entity) {
	var newList []blockchainRequest.Entity

	for _, req := range reqList {
		reqKey := string(req.Seal.Hash)
		if _, exists := u.reqPool[reqKey]; exists {
			delete(u.reqPool, reqKey)
			continue
		}
	}

	for _, req := range u.reqList {
		reqKey := string(req.Seal.Hash)
		if _, exists := u.reqPool[reqKey]; exists {
			newList = append(newList, req)
		}
	}

	u.reqList = newList
}

func (u *UniversalIdentificationApplication) Execute(
	req blockchainRequest.Entity,
	blk block.Entity,
	actIndex uint32,
) (result applicationResult.Entity, err error) {
	reqList := ActionList{}
	err = structSerializer.FromMFBytes(req.Data, &reqList)
	if err != nil {
		return
	}

	u.poolLock.Lock()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}

		u.poolLock.Unlock()
	}()

	for _, req := range reqList.Actions {
		reqKey := string(req.Seal.Hash)
		if _, exist := u.reqPool[reqKey]; exist {
			err = errors.New("request not exist")
			return
		}

		err = u.ledger.ExecuteAction(req.RequestAction, req.Data)
		if err != nil {
			return
		}
	}

	u.removeRequestFromPool(reqList.Actions)
	return
}

func (u *UniversalIdentificationApplication) Cancel(req blockchainRequest.Entity) (err error) {return}

func (u *UniversalIdentificationApplication) RequestsForBlock(_ block.Entity) (reqList []blockchainRequest.Entity, cnt uint32) {
	u.poolLock.Lock()

	actList := ActionList{}
	actList.Actions = u.reqList

	mt := merkleTree.Tree{}
	for _, req := range u.reqPool {
		mt.AddHash(req.Seal.Hash)
	}

	txRoot, _ := mt.Calculate()
	reqData, _ := structSerializer.ToMFBytes(actList)

	packedReq := blockchainRequest.Entity{
		EntityData: blockchainRequest.EntityData{
			RequestApplication: u.Name(),
			RequestAction:      "",
			Data:               reqData,
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

	u.poolLock.Unlock()
	return []blockchainRequest.Entity{packedReq}, 1
}

func (u *UniversalIdentificationApplication) Information() (info service.BasicInformation) {
	info.Name = u.Name()
	info.Description = "this is a universal identification application"

	info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
	info.Api.Address = ""
	info.Api.ApiList = []service.ApiInterface {}
	return
}

func (u *UniversalIdentificationApplication) SetChainInterface(_ chainStructure.IChainInterface) {}
func (u *UniversalIdentificationApplication) UnpackingActionsAsRequests(_ blockchainRequest.Entity) ([]blockchainRequest.Entity, error) {return nil, nil}
func (u *UniversalIdentificationApplication) GetActionAsRequest(req blockchainRequest.Entity) (newReq blockchainRequest.Entity) {

	return
}

func Load()  {}

func NewApplicationInterface(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver) (app chainStructure.IBlockchainExternalApplication) {
	uidApp := UniversalIdentificationApplication{}

	uidApp.ledger = uidLedger.NewLedger(kvDriver, sqlDriver)

	return &uidApp
}
