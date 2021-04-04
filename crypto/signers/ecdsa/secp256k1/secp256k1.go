/*
 * Copyright 2021 The SealABC Authors
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

package secp256k1

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/SealSC/SealABC/crypto/signers/signerCommon"
	"github.com/btcsuite/btcd/btcec"
)

const algorithmName = "secp256k1"

type keyPair struct {
	PrivateKey *btcec.PrivateKey
	PublicKey  *btcec.PublicKey
}

func (k keyPair)Type() string {
	return algorithmName
}

func (k keyPair)Sign(data []byte) (signature []byte, err error) {
	if k.PrivateKey == nil {
		return nil, errors.New("no private key")
	}

	sig, err := k.PrivateKey.Sign(data)
	if err != nil {
		return
	}


	return sig.Serialize(), nil
}

func (k keyPair)Verify(data []byte, signature []byte) (passed bool, err error) {
	if k.PublicKey == nil {
		return false, errors.New("no public key")
	}

	sig, err := btcec.ParseSignature(signature, btcec.S256())
	if err != nil {
		return
	}


	return sig.Verify(data, k.PublicKey), nil
}

func (k keyPair)RawKeyPair() (kp interface{}) {
	return nil
}

func (k keyPair)KeyPairData() (keyData []byte) {
	return nil
}

func (k keyPair)PublicKeyString() (key string) {
	keyBytes := k.PublicKeyBytes()
	return hex.EncodeToString(keyBytes)
}

func (k keyPair)PrivateKeyString() (key string) {
	keyBytes := k.PrivateKeyBytes()
	return hex.EncodeToString(keyBytes)
}

func (k keyPair)PublicKeyBytes() (key [] byte) {
	if k.PublicKey == nil {
		return
	}

	return k.PublicKey.SerializeCompressed()
}

func (k keyPair)PrivateKeyBytes() (key []byte) {
	if k.PrivateKey == nil {
		return
	}

	return k.PrivateKey.Serialize()
}

func (k keyPair)PublicKeyCompare(pub interface{}) (equal bool) {
	pubBytes, ok := pub.([]byte)
	if !ok {
		return false
	}

	return bytes.Equal(k.PublicKeyBytes(), pubBytes)
}

type keyGenerator struct {}

func (keyGenerator) Type() string {
	return algorithmName
}


func (keyGenerator) NewSigner(_ interface{}) (s signerCommon.ISigner, err error) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return
	}

	pub := priv.PubKey()

	s = &keyPair{
		PrivateKey: priv,
		PublicKey:  pub,
	}

	return s, nil
}

func (k *keyGenerator) FromRawPrivateKey(key interface{}) (s signerCommon.ISigner, err error) {
	keyBytes, ok := key.([]byte)
	if !ok {
		err = errors.New("only support bytes type key")
		return
	}
	if len(keyBytes) != btcec.PrivKeyBytesLen {
		err = errors.New("invalid key size")
		return
	}

	priv, pub := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)

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
	if len(keyBytes) != btcec.PubKeyBytesLenCompressed &&
		len(keyBytes) != btcec.PubKeyBytesLenUncompressed &&
		len(keyBytes) != btcec.PubKeyBytesLenHybrid{
		err = errors.New("invalid key size")
		return
	}

	pub, err := btcec.ParsePubKey(keyBytes, btcec.S256())
	if err != nil {
		return
	}

	s =  &keyPair{
		PublicKey: pub,
	}
	return
}

func (k *keyGenerator) FromKeyPairData(_ []byte) (signer signerCommon.ISigner, err error)  {
	err = errors.New("only support gen from key bytes")
	return
}

func (k *keyGenerator) FromRawKeyPair(keys interface{}) (s signerCommon.ISigner, err error) {
	err = errors.New("only support gen from key bytes")
	return
}

var SignerGenerator = &keyGenerator{}
