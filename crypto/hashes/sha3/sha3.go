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

package sha3

import (
    "SealABC/crypto/hashes/commonHash"
    "golang.org/x/crypto/sha3"
)

const (
    sha3_256 = "sha3_256"
    sha3_512 = "sha3_512"
    keccak_256 = "keccak_256"
    keccak_512 = "keccak_512"
)

func Load() {
    commonHash.RegisterHashMethod(sha3_256, sha3.New256)
    commonHash.RegisterHashMethod(sha3_512, sha3.New512)
    commonHash.RegisterHashMethod(keccak_256, sha3.NewLegacyKeccak256)
    commonHash.RegisterHashMethod(keccak_512, sha3.NewLegacyKeccak512)
}

var Sha256 = &commonHash.CommonHash{sha3_256}
var Sha512 = &commonHash.CommonHash{sha3_512}
var Keccak256 = &commonHash.CommonHash{keccak_256}
var Keccak512 = &commonHash.CommonHash{keccak_512}
