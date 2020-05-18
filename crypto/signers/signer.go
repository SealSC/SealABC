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

type KeyPair struct{
    PrivateKey  interface{}
    PublicKey   interface{}
}

type ISigner interface {
    Name() string
    Sign(data []byte) (signature []byte)
    Verify(data []byte, signature []byte) (passed bool)
    RawKeyPair() (kp KeyPair)
    KeyPairData() (keyData []byte)

    PublicKeyString() (key string)
    PrivateKeyString() (key string)

    PublicKeyBytes() (key [] byte)
    PrivateKeyBytes() (key []byte)

    PublicKeyCompare(k interface{}) (equal bool)
}

type ISignerGenerator interface {
    NewSigner(param interface{}) (signer ISigner, err error)
    FromKeyPairData(kpData []byte) (kp ISigner, err error)
    FromRawPrivateKey(key interface{}) (signer ISigner, err error)
    FromRawPublicKey(key interface{}) (signer ISigner, err error)
    FromRawKeyPair(kp KeyPair) (signer ISigner, err error)
}