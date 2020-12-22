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

package ed25519

import (
    "github.com/SealSC/SealABC/crypto/signers/signerCommon"
    "bytes"
    "crypto/ed25519"
    "encoding/hex"
    "encoding/json"
    "errors"
)

const algorithmName = "ED25519"

type keyPair struct {
    PrivateKey ed25519.PrivateKey
    PublicKey  ed25519.PublicKey
}

func (k keyPair) Type() string {
    return algorithmName
}

func (k keyPair) Verify(data []byte, signature []byte) (passed bool) {
    if len(k.PublicKey) != ed25519.PublicKeySize {
        return false
    }
    passed = ed25519.Verify(k.PublicKey, data, signature)
    return
}

func (k keyPair) Sign(data []byte) (signature []byte)  {
    if len(k.PrivateKey) != ed25519.PrivateKeySize {
        return
    }

    signature = ed25519.Sign(k.PrivateKey, data)
    return
}

func (k keyPair)KeyPairData() (keyData []byte) {
    keyData, _ =json.Marshal(k)
    return
}

func (k keyPair)RawKeyPair() (kp interface{}) {
    return signerCommon.KeyPair {
        PrivateKey: append([]byte{}, k.PrivateKey...),
        PublicKey:  append([]byte{}, k.PublicKey...),
    }
}

func (k keyPair) PublicKeyBytes() (key [] byte) {
    return append([]byte{}, k.PublicKey...)
}
func (k keyPair) PrivateKeyBytes() (key []byte) {
    return append([]byte{}, k.PrivateKey...)
}

func (k keyPair) PublicKeyString() (key string)  {
    return hex.EncodeToString(k.PublicKey)
}

func (k keyPair) PrivateKeyString() (key string)  {
    return hex.EncodeToString(k.PrivateKey)
}

func (k keyPair) PublicKeyCompare(key interface{}) (equal bool) {
    keyBytes, ok := key.([]byte)
    if !ok {
        return false
    }

    return bytes.Equal(k.PublicKey, keyBytes)
}

func calcPublicKey(priv []byte) (pub []byte) {
    pub = ed25519.PrivateKey(priv).Public().(ed25519.PublicKey)
    return
}

type keyGenerator struct {}

func (keyGenerator) Type() string {
    return algorithmName
}

func (keyGenerator) NewSigner(_ interface{}) (s signerCommon.ISigner, err error) {
    pub, priv, err := ed25519.GenerateKey(nil)
    if err != nil {
        return
    }

    //checking the key length in case to avoid potentially golang interface incompatible caused private key expose
    if len(priv) != ed25519.PrivateKeySize || len(pub) != ed25519.PublicKeySize {
        err = errors.New("invalid key size")
        return
    }

    s = &keyPair{
        PrivateKey: priv,
        PublicKey:  pub,
    }

    return
}

func (k *keyGenerator) FromRawPrivateKey(key interface{}) (s signerCommon.ISigner, err error) {
    keyBytes, ok := key.([]byte)
    if !ok {
        err = errors.New("only support bytes type key")
        return
    }
    if len(keyBytes) != ed25519.PrivateKeySize {
        err = errors.New("invalid key size")
        return
    }

    priv := append([]byte{}, keyBytes...)
    pub := calcPublicKey(priv)

    s = &keyPair{
        PrivateKey: priv,
        PublicKey:  pub,
    }

    return
}

func (k *keyGenerator) FromRawPublicKey(key interface{}) (s signerCommon.ISigner, err error) {
    keyBytes, ok := key.([]byte)
    if !ok {
        err = errors.New("only support bytes type key")
        return
    }
    if len(keyBytes) != ed25519.PublicKeySize {
        err = errors.New("invalid key size")
        return
    }

    s =  &keyPair{
        PublicKey: keyBytes,
    }
    return
}

func (k *keyGenerator) FromKeyPairData(kpData []byte) (signer signerCommon.ISigner, err error)  {
    newSigner := keyPair{}
    err = json.Unmarshal(kpData, &newSigner)
    if err != nil {
        return
    }

    signer = &newSigner
    return
}

func (k *keyGenerator) FromRawKeyPair(keys interface{}) (s signerCommon.ISigner, err error) {
    newSigner := keyPair{}

    kp := keys.(signerCommon.KeyPair)
    newSigner.PrivateKey = append([]byte{}, kp.PrivateKey.([]byte)...)
    newSigner.PublicKey = append([]byte{}, kp.PublicKey.([]byte)...)

    s = &newSigner
    return
}

var SignerGenerator = &keyGenerator{}

