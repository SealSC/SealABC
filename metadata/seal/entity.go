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
    "github.com/SealSC/SealABC/crypto"
    "github.com/SealSC/SealABC/crypto/hashes"
    "github.com/SealSC/SealABC/crypto/signers"
    "bytes"
    "encoding/hex"
    "errors"
)

type Entity struct {
    Hash            []byte
    Signature       []byte
    SignerPublicKey []byte
    SignerAlgorithm string
}

func (e *Entity) IsPureEmpty() bool {
    return  len(e.Hash) == 0 &&
            len(e.Signature) == 0 &&
            len(e.SignerPublicKey) == 0 &&
            e.SignerAlgorithm == ""
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

func (e *Entity) Sign(orgData []byte, tools crypto.Tools, privateKey interface{}) (err error) {
    s, err := tools.SignerGenerator.FromRawPrivateKey(privateKey)
    if err != nil {
        return
    }

    e.Hash = tools.HashCalculator.Sum(orgData)
    e.Signature = s.Sign(e.Hash)

    e.SignerAlgorithm = tools.SignerGenerator.Type()
    e.SignerPublicKey = s.PublicKeyBytes()

    return
}

func (e Entity) Verify(orgData []byte, hashCalc hashes.IHashCalculator) (passed bool, err error) {
    passed = false

    hash := hashCalc.Sum(orgData)
    if !bytes.Equal(hash, e.Hash) {
        err = errors.New("hash not equal: " + hex.EncodeToString(hash) + " vs " + hex.EncodeToString(e.Hash) )
        return
    }

    signerGen := signers.SignerGeneratorByAlgorithmType(e.SignerAlgorithm)
    if signerGen == nil {
        err = errors.New("unsupported signature algorithm:" + e.SignerAlgorithm)
        return
    }

    signer, err := signerGen.FromRawPublicKey(e.SignerPublicKey)
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
