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

package smartAssetsSQLTables

import (
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsLedger"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

type ISQLRows interface {
	simpleSQLDatabase.IRows
	Insert(tx smartAssetsLedger.Transaction, blk block.Entity)
}

func Load()  {
	AddressList.load()
	ContractCall.load()
	Contract.load()
	Transaction.load()
	Transfer.load()
}
