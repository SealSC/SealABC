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

package pbkdf2

import (
    "SealABC/crypto/hashes"
    "SealABC/crypto/hashes/sha3"
    "encoding/json"
    "errors"
    "golang.org/x/crypto/bcrypt"
    "golang.org/x/crypto/pbkdf2"
)

type Param struct {
    Iter                int
    KeyHashAlgorithm    string
    PasswordHashCost    int
}

type keyGenerator struct {}
var Generator keyGenerator

var defaultParam = Param {
    Iter:             10000,
    KeyHashAlgorithm: sha3.Keccak512.Name(),
    PasswordHashCost: bcrypt.DefaultCost,

}

func (p keyGenerator) Name() string {
    return "PBKDF2"
}

func (p keyGenerator) NewKey(password []byte, keyLen int) (key []byte, salt []byte, paramData []byte, err error) {
    salt, err = bcrypt.GenerateFromPassword(password, defaultParam.PasswordHashCost)
    if err != nil {
        return
    }

    key = pbkdf2.Key(password, salt, defaultParam.Iter, keyLen, hashes.AllHash[defaultParam.KeyHashAlgorithm].OriginalHash())
    param := Param {
        Iter:             defaultParam.Iter,
        KeyHashAlgorithm: defaultParam.KeyHashAlgorithm,
        PasswordHashCost: defaultParam.PasswordHashCost,
    }

    paramData, err = json.Marshal(param)
    return
}

func (p keyGenerator) NewKeyWithParam(password []byte, keyLen int, param []byte) (key []byte, salt []byte, err error)  {
    customParam := Param{}
    err = json.Unmarshal(param, &customParam)
    if err != nil {
        return
    }

    salt, err = bcrypt.GenerateFromPassword(password, customParam.PasswordHashCost)
    if err != nil {
        return
    }

    hashCalc, supported := hashes.AllHash[customParam.KeyHashAlgorithm]
    if !supported {
        err = errors.New("not supported hash: " + customParam.KeyHashAlgorithm)
        return
    }
    key = pbkdf2.Key(password, salt, customParam.Iter, keyLen, hashCalc.OriginalHash())
    return
}

func (p keyGenerator) RebuildKey(password []byte, keyLen int, salt []byte, param []byte) (key []byte, err error) {
    customParam := Param{}
    err = json.Unmarshal(param, &customParam)
    if err != nil {
        return
    }

    hashCalc, supported := hashes.AllHash[customParam.KeyHashAlgorithm]
    if !supported {
        err = errors.New("not supported hash: " + customParam.KeyHashAlgorithm)
        return
    }
    key = pbkdf2.Key(password, salt, customParam.Iter, keyLen, hashCalc.OriginalHash())
    return
}
