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

package block

import (
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/metadata/seal"
)

type Header struct {
	Version          string
	Height           uint64
	PrevBlock        []byte
	TransactionsRoot []byte
	StateRoot        map[string][]byte
	Timestamp        uint64
}

type Body struct {
	RequestsCount int
	Requests      []blockchainRequest.Entity
}

type EntityData struct {
	Header Header
	Body   Body
}

type Entity struct {
	EntityData

	Seal      seal.Entity
	BlankSeal seal.Entity
}
