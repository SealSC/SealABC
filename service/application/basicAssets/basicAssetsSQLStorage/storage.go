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

package basicAssetsSQLStorage

import (
	"errors"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/service/application/basicAssets/basicAssetsLedger"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

var QueryTypes struct {
	AssetsList enum.Element
	Assets     enum.Element

	Transfer             enum.Element
	TransferList         enum.Element
	TransfersUnderAssets enum.Element
	BalancesUnderAssets  enum.Element

	AddressList         enum.Element
	AddressActionRecord enum.Element
	AddressBalanceList  enum.Element

	SellingHistory enum.Element
}

type queryHandler func([]string) (interface{}, error)
type Storage struct {
	queryHandlers map[string]queryHandler
	Driver        simpleSQLDatabase.IDriver
}

func Load() {
	enum.SimpleBuild(&QueryTypes)
}

func NewStorage(sqlDriver simpleSQLDatabase.IDriver) (s *Storage) {
	s = &Storage{
		Driver: sqlDriver,
	}

	s.queryHandlers = map[string]queryHandler{
		QueryTypes.AssetsList.String():           s.GetAssetsList,
		QueryTypes.TransferList.String():         s.GetTransferList,
		QueryTypes.Transfer.String():             s.GetTransfer,
		QueryTypes.Assets.String():               s.GetAssets,
		QueryTypes.AddressList.String():          s.GetAddressesList,
		QueryTypes.TransfersUnderAssets.String(): s.GetTransfersUnderAssets,
		QueryTypes.BalancesUnderAssets.String():  s.GetBalancesUnderAssetsList,

		QueryTypes.AddressActionRecord.String(): s.GetAddressActionRecord,
		QueryTypes.AddressBalanceList.String():  s.GetAddressBalance,

		QueryTypes.SellingHistory.String(): s.GetSellingHistory,
	}
	return
}

func (s *Storage) DoQuery(queryReq basicAssetsLedger.QueryRequest) (result interface{}, err error) {
	if handler, exists := s.queryHandlers[queryReq.QueryType]; !exists {
		err = errors.New("no such query handler: " + queryReq.QueryType)
		return
	} else {
		return handler(queryReq.Parameter)
	}
}
