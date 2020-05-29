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

package smartAssetsLedger

import "SealABC/dataStructure/enum"

var QueryTypes struct {
	BaseAssets   enum.Element
	Balance      enum.Element
	Transaction  enum.Element
	OffChainCall enum.Element
}

var QueryParameterFields struct{
	Address enum.Element
	TxHash  enum.Element
	Data    enum.Element
}

type QueryRequest struct {
	QueryType string
	Parameter map[string] string
}
