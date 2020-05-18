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
	"SealABC/crypto/hashes"
	"SealABC/crypto/hashes/sha3"
	"SealABC/metadata/seal"
	"bytes"
)

func (s SealAccount) Sign(data []byte) []byte {
	return s.Signer.Sign(data)
}

func (s SealAccount) Seal(data []byte, hashCalculator hashes.IHashCalculator) seal.Entity {
	if hashCalculator == nil {
		hashCalculator = sha3.Sha256
	}

	dataHash := hashCalculator.Sum(data)
	sig := s.Signer.Sign(dataHash)

	return seal.Entity{
		Hash:            dataHash,
		Signature:       sig,
		SignerPublicKey: s.Signer.PrivateKeyBytes(),
	}
}

func (s SealAccount) VerifySignature(data []byte, sig []byte) bool {
	return s.Signer.Verify(data, sig)
}

func (s SealAccount) VerifySeal(data []byte, sl seal.Entity, hashCalculator hashes.IHashCalculator) bool {
	if !bytes.Equal(s.Signer.PublicKeyBytes(), sl.SignerPublicKey) {
		return false
	}

	if hashCalculator == nil {
		hashCalculator = sha3.Sha256
	}

	dataHash := hashCalculator.Sum(data)
	if !bytes.Equal(dataHash, sl.Hash) {
		return false
	}

	return s.Signer.Verify(dataHash, sl.Signature)
}
