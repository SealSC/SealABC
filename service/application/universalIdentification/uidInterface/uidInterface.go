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
	"github.com/SealSC/SealABC/metadata/applicationResult"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
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
	return
}

func (u *UniversalIdentificationApplication) Query(req []byte) (result interface{}, err error) {
	return
}

func (u *UniversalIdentificationApplication) PreExecute(req blockchainRequest.Entity, _ block.Entity) (result []byte, err error) {
	return
}

func (u *UniversalIdentificationApplication) Execute(
	req blockchainRequest.Entity,
	blk block.Entity,
	actIndex uint32,
) (result applicationResult.Entity, err error) {return}

func (u *UniversalIdentificationApplication) Cancel(req blockchainRequest.Entity) (err error) {return}

func (u *UniversalIdentificationApplication) RequestsForBlock(_ block.Entity) (reqList []blockchainRequest.Entity, cnt uint32) {return}

func (u *UniversalIdentificationApplication) Information() (info service.BasicInformation) {
	info.Name = u.Name()
	info.Description = "this is a universal identification application"

	info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
	info.Api.Address = ""
	info.Api.ApiList = []service.ApiInterface {}
	return
}

func (u *UniversalIdentificationApplication) SetBlockchainService(_ interface{}) {}
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
