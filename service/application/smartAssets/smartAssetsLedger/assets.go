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

import "SealABC/metadata/seal"

const baseAssetsPrefix = "base-assets-"

type BaseAssetsData struct {
	Name        string
	Symbol      string
	Supply      string
	Increasable bool
	Creator     string
}

type BaseAssets struct {
	BaseAssetsData

	IssuedSeal seal.Entity
	MetaSeal   seal.Entity
}
