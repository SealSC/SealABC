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

package basicAssetsLedger

import "SealABC/dataStructure/enum"

var QueryTypes struct {
    Assets      enum.Element
    AllAssets   enum.Element
    UnspentList enum.Element
    Transaction enum.Element
}

type AssetsList struct {
    List []Assets
}

type UnspentUnderAssets struct {
    Assets      Assets
    UnspentList []Unspent
}

type UnspentList struct {
    List map[string] *UnspentUnderAssets
}

type UnspentQueryParameter struct {
    Address []byte
    Assets  []byte
}

type QueryRequest struct {
    DBType    string
    QueryType string
    Parameter []string
}

