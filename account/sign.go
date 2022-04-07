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

package account

import (
	"bytes"
	"github.com/SealSC/SealABC/crypto/hashes"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/metadata/seal"
)

func (s SealAccount) Sign(data []byte) ([]byte, error) {
	return s.Signer.Sign(data)
}

func (s SealAccount) Seal(data []byte, hashCalculator hashes.IHashCalculator) (se seal.Entity, err error) {
	if hashCalculator == nil {
		hashCalculator = sha3.Sha256
	}

	dataHash := hashCalculator.Sum(data)
	sig, err := s.Signer.Sign(dataHash)
	if err != nil {
		return
	}

	return seal.Entity{
		Hash:            dataHash,
		Signature:       sig,
		SignerPublicKey: s.Signer.PrivateKeyBytes(),
	}, nil
}

func (s SealAccount) VerifySignature(data []byte, sig []byte) (bool, error) {
	return s.Signer.Verify(data, sig)
}

func (s SealAccount) VerifySeal(data []byte, sl seal.Entity, hashCalculator hashes.IHashCalculator) (bool, error) {
	if !bytes.Equal(s.Signer.PublicKeyBytes(), sl.SignerPublicKey) {
		return false, nil
	}

	if hashCalculator == nil {
		hashCalculator = sha3.Sha256
	}

	dataHash := hashCalculator.Sum(data)
	if !bytes.Equal(dataHash, sl.Hash) {
		return false, nil
	}

	return s.Signer.Verify(dataHash, sl.Signature)
}
