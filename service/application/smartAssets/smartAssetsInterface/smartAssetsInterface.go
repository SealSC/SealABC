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

package smartAssetsInterface

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/crypto"
	"SealABC/metadata/applicationResult"
	"SealABC/metadata/block"
	"SealABC/metadata/blockchainRequest"
	"SealABC/metadata/seal"
	"SealABC/service"
	"SealABC/service/application/smartAssets/smartAssetsLedger"
	"SealABC/service/system/blockchain/chainStructure"
	"SealABC/storage/db/dbInterface/kvDatabase"
	"SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

type SmartAssetsApplication struct {
	ledger *smartAssetsLedger.Ledger
}

func (s *SmartAssetsApplication) Name() (name string) {
	return "Smart Assets"
}

func (s *SmartAssetsApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
	err = s.ledger.AddTx(req)
	return nil, err
}

func (s *SmartAssetsApplication) Query(req string) (result interface{}, err error) {
	return
}

func (s *SmartAssetsApplication) PreExecute(req blockchainRequest.Entity, blk block.Entity) (result []byte, err error) {
	var txList smartAssetsLedger.TransactionList
	err = structSerializer.FromMFBytes(req.Data, &txList)
	if err != nil {
		return
	}

	return s.ledger.PreExecute(txList, blk)
}

func (s *SmartAssetsApplication) Execute(
	req blockchainRequest.Entity,
	blk block.Entity,
	actIndex uint32,
) (result applicationResult.Entity, err error) {
	txList := smartAssetsLedger.TransactionList{}
	err = structSerializer.FromMFBytes(req.Data, &txList)
	if err != nil {
		return
	}


	_, err = s.ledger.Execute(txList, blk)

	return
}

func (s *SmartAssetsApplication) Cancel(req blockchainRequest.Entity) (err error) {
	return
}

func (s *SmartAssetsApplication) RequestsForBlock(blk block.Entity) (reqList []blockchainRequest.Entity, cnt uint32) {
	txList, cnt := s.ledger.GetTransactionsFromPool(blk)
	if cnt == 0 {
		return
	}

	reqData, _ := structSerializer.ToMFBytes(txList)
	packReq := blockchainRequest.Entity{
		EntityData: blockchainRequest.EntityData{
			RequestApplication: s.Name(),
			RequestAction:      "",
			Data:               reqData,
			QueryString:        "",
		},
		Packed:     true,
		Seal:       seal.Entity{},
	}

	return []blockchainRequest.Entity{packReq}, 1
}

func (s *SmartAssetsApplication) Information() (info service.BasicInformation) {
	info.Name = s.Name()
	info.Description = "this is a smart assets application based on a balance mode ledger and EVM supported"

	info.Api.Protocol = service.ApiProtocols.INTERNAL.String()
	info.Api.Address = ""
	info.Api.ApiList = []service.ApiInterface {}
	return
}

func (s *SmartAssetsApplication) SetBlockchainService(bs interface{}) {
	chain, ok := bs.(*chainStructure.Blockchain)
	if !ok {
		return
	}
	s.ledger.SetChain(chain)
}

func Load()  {}

func NewApplicationInterface(
	kvDriver kvDatabase.IDriver,
	sqlDriver simpleSQLDatabase.IDriver,
	tools crypto.Tools,
	assets smartAssetsLedger.BaseAssetsData,
	) (app chainStructure.IBlockchainExternalApplication, err error) {
	sa := SmartAssetsApplication{}

	sa.ledger = &smartAssetsLedger.Ledger {
		CryptoTools: tools,
		Storage:     kvDriver,
	}

	err = sa.ledger.LoadGenesisAssets(nil, assets)
	if err != nil {
		return
	}

	app = &sa
	return
}
