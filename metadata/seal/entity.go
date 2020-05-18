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

package seal

import (
    "SealABC/crypto"
    "bytes"
    "encoding/hex"
    "errors"
)

type Entity struct {
    Hash            []byte
    Signature       []byte
    SignerPublicKey []byte
}

func (e *Entity) HexHash() string {
    return hex.EncodeToString(e.Hash)
}

func (e *Entity) HexPublicKey() string {
    return hex.EncodeToString(e.SignerPublicKey)
}

func (e *Entity) HexSignature() string {
    return hex.EncodeToString(e.Signature)
}

func (e *Entity) Sign(orgData []byte, tools crypto.Tools, privateKey []byte) (err error) {
    s, err := tools.SignerGenerator.FromRawPrivateKey(privateKey)
    if err != nil {
        return
    }

    e.Hash = tools.HashCalculator.Sum(orgData)
    e.Signature = s.Sign(e.Hash)
    e.SignerPublicKey = s.PublicKeyBytes()

    return
}

func (e Entity) Verify(orgData []byte, tools crypto.Tools) (passed bool, err error) {
    passed = false

    hash := tools.HashCalculator.Sum(orgData)
    if !bytes.Equal(hash, e.Hash) {
        err = errors.New("hash not equal")
        return
    }

    signer, err := tools.SignerGenerator.FromRawPublicKey(e.SignerPublicKey)
    if err != nil {
        err = errors.New("invalid seal's public key")
        return 
    }

    passed = signer.Verify(hash, e.Signature)
    if !passed {
        err = errors.New("invalid seal signature")
    }
    
    return 
}
