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

package kvDatabase

type IDriver interface {
	Close()

	Put(kv KVItem) (err error)
	Get(k []byte) (kv KVItem, err error)
	Delete(k []byte) (err error)
	Check(k []byte) (exists bool, err error)

	BatchPut(kvList []KVItem) (err error)
	BatchGet(kList [][]byte) (kvList []KVItem, err error)
	BatchDelete(kList [][]byte) (err error)
	BatchCheck(kList [][]byte) (kvList []KVItem, err error)

	NewBatch() Batch
	BatchWrite(Batch) (err error)

	Traversal(condition []byte) (kvList []KVItem)

	Stat() (state interface{}, err error)
}
