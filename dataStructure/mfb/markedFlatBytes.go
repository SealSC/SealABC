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

package mfb

import (
    "encoding/binary"
    "bytes"
)

const mark_bytes_count = 4

type MarkedFlatBytes []byte

func (m *MarkedFlatBytes)FromByteSlice(bufferList [][]byte) {
    //buffer := bytes.Buffer{}

    for _, b := range bufferList {
        var lenBuf = make([]byte, mark_bytes_count)
        binary.BigEndian.PutUint32(lenBuf, uint32(len(b)))

        //buffer.Write(lenBuf)
        //buffer.Write(b)
        *m = append(*m, lenBuf...)
        *m = append(*m, b...)
    }

    //*m = buffer.Bytes()
    return
}

func (m MarkedFlatBytes)ToByteSlice() (bufferList [][]byte, err error) {
    if len(m) < mark_bytes_count {
        return
    }

    buffer := bytes.Buffer{}
    buffer.Write(m)

    for {
        if buffer.Len() <= 0 {
            break
        }

        lenBytes := make([]byte, mark_bytes_count)
        _, err = buffer.Read(lenBytes)
        if err != nil {
            return
        }

        elLength := int(binary.BigEndian.Uint32(lenBytes))

        if elLength > buffer.Len() {
            break
        }

        bufferList = append(bufferList, buffer.Next(elLength))
    }

    return
}
