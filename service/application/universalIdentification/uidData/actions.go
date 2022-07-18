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

package uidData

import (
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/metadata/seal"
)

type UIDCreation struct {
	UID  UniversalIdentification
	Seal seal.Entity
}

type UIDAppendKeysData struct {
	Identification string
	Keys           []UIDKey
	NewUIDSeal     seal.Entity
}

type UIDAppendKeys struct {
	UIDAppendKeysData

	Seal seal.Entity
}

type UIDKeyWithIndex struct {
	KeyIndex int
	UIDKey
}

type UIDUpdateKeysData struct {
	Identification string
	NewKeys        []UIDKeyWithIndex
	NewUIDSeal     seal.Entity
}

type UIDUpdateKeys struct {
	UIDUpdateKeysData

	Seal seal.Entity
}

type UIDQuery struct {
	PublicKey []byte
	Namespace string
}

type QueryResult struct {
	UIDList []UniversalIdentification
}

var UIDActionTypes struct {
	Create enum.Element
	Append enum.Element
	Update enum.Element
}
