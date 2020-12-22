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

package ciphers

import (
    "github.com/SealSC/SealABC/crypto/ciphers/aes"
    "github.com/SealSC/SealABC/crypto/ciphers/cipherCommon"
)

var ciphers = map[string] ICipher {
    aes.Cipher.Type(): aes.Cipher,
}

type ICipher interface {
    Type() string
    Encrypt(plainText []byte, key []byte, param interface{}) (result cipherCommon.EncryptedData, err error)
    Decrypt(cipherText cipherCommon.EncryptedData, key []byte, param interface{}) (plaintext []byte, err error)
}

func CipherByAlgorithmType(cType string) ICipher {
    return ciphers[cType]
}

