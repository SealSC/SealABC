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

package oracleInterface

import (
	"crypto/md5"
	"crypto/ed25519"
	"encoding/binary"
	"time"
)

func MD5(bs []byte) []byte {
	hash := md5.New()
	hash.Write(bs)
	sum := hash.Sum(nil)
	return sum
}
func randomAddr(data []byte) ([]byte, []byte, []byte) {
	var sig = make([]byte, ed25519.SeedSize)
	binary.BigEndian.PutUint64(sig, uint64(time.Now().UnixNano()))
	key := ed25519.NewKeyFromSeed(sig)
	pub := key[32:]
	hash := append(MD5(data), MD5(key)...)
	return pub, sig, hash
}
