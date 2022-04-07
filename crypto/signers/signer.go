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

package signers

import (
	"github.com/SealSC/SealABC/crypto/signers/ecdsa/secp256k1"
	"github.com/SealSC/SealABC/crypto/signers/ed25519"
	"github.com/SealSC/SealABC/crypto/signers/signerCommon"
	"github.com/SealSC/SealABC/crypto/signers/sm2"
)

var signerGenerators = map[string]ISignerGenerator{
	ed25519.SignerGenerator.Type():   ed25519.SignerGenerator,
	secp256k1.SignerGenerator.Type(): secp256k1.SignerGenerator,
	sm2.SignerGenerator.Type():       sm2.SignerGenerator,
}

type ISignerGenerator interface {
	Type() string
	NewSigner(param interface{}) (signer signerCommon.ISigner, err error)
	FromKeyPairData(kpData []byte) (kp signerCommon.ISigner, err error)
	FromRawPrivateKey(key interface{}) (signer signerCommon.ISigner, err error)
	FromRawPublicKey(key interface{}) (signer signerCommon.ISigner, err error)
	FromRawKeyPair(kp interface{}) (signer signerCommon.ISigner, err error)
}

func SignerGeneratorByAlgorithmType(sType string) ISignerGenerator {
	return signerGenerators[sType]
}
