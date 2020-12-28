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
	"github.com/SealSC/SealABC/metadata/seal"
)

type UIDCreateTransaction struct {
	Keys []UIDKey
	Seal seal.Entity
}

type UIDAppendKeysTransaction struct {
	UID  string
	Keys []UIDKey
	Seal seal.Entity
}

type UIDKeyToUpdate struct {
	KeyIndex int
	UIDKey
}

type UIDUpdateKeysTransaction struct {
	UID     string
	NewKeys []UIDKeyToUpdate
	Seal    seal.Entity
}
