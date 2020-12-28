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
)

type UniversalIdentification struct {

}

func (u *UniversalIdentification) Name() (name string) {
	return "Universal Identification"
}

func (u *UniversalIdentification) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
	return
}

func (u *UniversalIdentification) Query(req []byte) (result interface{}, err error) {
	return
}

func (u *UniversalIdentification) PreExecute(req blockchainRequest.Entity, _ block.Entity) (result []byte, err error) {
	return
}

func (u *UniversalIdentification) Execute(
	req blockchainRequest.Entity,
	blk block.Entity,
	actIndex uint32,
) (result applicationResult.Entity, err error) {return}

func (u *UniversalIdentification) Cancel(req blockchainRequest.Entity) (err error) {return}

func (u *UniversalIdentification) RequestsForBlock(_ block.Entity) (reqList []blockchainRequest.Entity, cnt uint32) {return}

func (u *UniversalIdentification) Information() (info service.BasicInformation) {
	info.Name = u.Name()
	info.Description = "this is a universal identification application"

	info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
	info.Api.Address = ""
	info.Api.ApiList = []service.ApiInterface {}
	return
}

func (u *UniversalIdentification) SetBlockchainService(_ interface{}) {}
func (u *UniversalIdentification) UnpackingActionsAsRequests(_ blockchainRequest.Entity) ([]blockchainRequest.Entity, error) {return nil, nil}
func (u *UniversalIdentification) GetActionAsRequest(req blockchainRequest.Entity) (newReq blockchainRequest.Entity) {
	return
}

func Load()  {
	uidLedger.Load()
}

func NewApplicationInterface(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver) (app chainStructure.IBlockchainExternalApplication) {
	return &UniversalIdentification{}
}
