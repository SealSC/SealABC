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

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
)

func (l *Ledger) storeCopyright(assets Assets, owner []byte) error {
	copyright := Copyright{
		Assets: assets,
		Owner:  owner,
	}

	key := l.buildCopyrightKey(assets.getUniqueHash())
	data, _ := json.Marshal(copyright)
	return l.Storage.Put(kvDatabase.KVItem{
		Key:  key,
		Data: data,
	})
}

func (l *Ledger) getCopyright(assetsHash []byte) (copyright Copyright, err error) {
	key := l.buildCopyrightKey(assetsHash)
	data, err := l.Storage.Get(key)
	if err != nil {
		return
	}

	if !data.Exists {
		err = errors.New("no such copyright")
		return
	}

	err = json.Unmarshal(data.Data, &copyright)
	return
}

func (l *Ledger) copyrightOwnerTransfer(assetsHash []byte, owner []byte) (err error) {
	copyright, err := l.getCopyright(assetsHash)
	if err != nil {
		return err
	}

	if !bytes.Equal(copyright.Owner, owner) {
		err = l.storeCopyright(copyright.Assets, owner)
	}

	return
}
