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

package hashes

import (
    "github.com/SealSC/SealABC/crypto/hashes/sha3"
    "hash"
)

var AllHash = map[string] IHashCalculator {
    sha3.Sha256.Name(): sha3.Sha256,
    sha3.Sha512.Name(): sha3.Sha512,
    sha3.Keccak256.Name(): sha3.Keccak256,
    sha3.Keccak512.Name(): sha3.Keccak512,
}

type IHashCalculator interface {
    Name() string
    Sum([]byte) []byte
    HMAC([]byte, []byte) []byte
    OriginalHash() func() hash.Hash
}

func Load()  {
    sha3.Load()
}
