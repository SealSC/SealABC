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

package block

import (
    "github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
    "github.com/SealSC/SealABC/crypto"
)

func (e *Entity) Sign(tools crypto.Tools, privateKey []byte) (err error) {
    blockHeaderBytes, _ := structSerializer.ToMFBytes(e.EntityData.Header)
    err = e.Seal.Sign(blockHeaderBytes, tools, privateKey)
    return
}

func (e *Entity) Verify(tools crypto.Tools) (passed bool, err error) {
    passed = false
    blockHeaderBytes, _ := structSerializer.ToMFBytes(e.EntityData.Header)

    passed, err = e.Seal.Verify(blockHeaderBytes, tools.HashCalculator)
    if err != nil {
        return
    }

    return
}
