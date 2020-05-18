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

package levelDB

import (
    "SealABC/storage/db/dbInterface/kvDatabase"
)

func (l *levelDBDriver)Put(kv kvDatabase.KVItem) (err error) {
    err = l.db.Put(kv.Key, kv.Data, nil)
    return
}

func (l *levelDBDriver)Get(k []byte) (kv kvDatabase.KVItem, err error) {
    v, err := l.db.Get(k, nil)
    if err != nil {
        return
    }

    kv.Key = k
    kv.Data = v
    kv.Exists = true
    return
}

func (l *levelDBDriver)Delete(k []byte) (err error) {
    err = l.db.Delete(k, nil)
    return
}

func (l *levelDBDriver)Check(k []byte) (exists bool, err error) {
    exists, err = l.db.Has(k, nil)
    return
}
