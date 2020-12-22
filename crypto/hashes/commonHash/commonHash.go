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

package commonHash

import (
	"github.com/SealSC/SealABC/log"
	"crypto/hmac"
	"hash"
)

type hashMethod func() hash.Hash

var hashMethodMap = map[string] func() hash.Hash {}

func calcHash(method hashMethod, data []byte) (hashBytes []byte) {
	newHash := method()
	_, err := newHash.Write(data)
	if err != nil {
		log.Log.Error("hash error: ", err.Error())
	}

	return newHash.Sum(nil)
}

func calcHMAC(method hashMethod, data [] byte, key []byte) []byte {
	hm := hmac.New(method, key)
	hm.Write(data)

	return hm.Sum(nil)
}

type CommonHash struct { HashType string }

func (c CommonHash) Name() string {
	return c.HashType
}

func (c CommonHash)Sum(data []byte) []byte {
	return calcHash(hashMethodMap[c.HashType], data)
}

func (c CommonHash) HMAC(data []byte, key []byte) []byte {
	return calcHMAC(hashMethodMap[c.HashType], data, key)
}

func (c CommonHash) OriginalHash() func() hash.Hash {
	return hashMethodMap[c.HashType]
}

func RegisterHashMethod(key string, method hashMethod) {
	if _, exist := hashMethodMap[key]; exist {
		return
	}

	hashMethodMap[key] = method
}

