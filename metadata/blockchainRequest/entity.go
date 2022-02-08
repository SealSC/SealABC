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

package blockchainRequest

import (
	"github.com/SealSC/SealABC/metadata/seal"
)

type EntityData struct {
	RequestApplication string
	RequestAction      string
	Data               []byte
	QueryString        string
}

type Entity struct {
	EntityData

	Packed      bool
	PackedCount uint32
	Seal        seal.Entity
	From        RequestFrom
}

func (e *Entity) SetFromIfNull(remote bool) {
	if e.From != 0 {
		return
	}
	if remote {
		e.From = RequestFromRemote
	} else {
		e.From = RequestFromAPI
	}
}

func (e *Entity) IsFromRemote() bool {
	return e.From == RequestFromRemote
}
func (e *Entity) IsFromAPI() bool {
	return e.From == RequestFromAPI
}

func EntityFromRemote() *Entity {
	e := &Entity{}
	e.SetFromIfNull(true)
	return e
}
func EntityFromAPI() *Entity {
	e := &Entity{}
	e.SetFromIfNull(false)
	return e
}

type RequestFrom int

const (
	RequestFromAPI    RequestFrom = 1
	RequestFromRemote RequestFrom = 2
)
