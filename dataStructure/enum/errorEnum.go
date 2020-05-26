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

package enum

import (
    "reflect"
    "strconv"
)

const (
    errorCodeTag    = "code"
    errorMessageTag = "msg"
)

type ErrorElement struct {
    code    int64
    name    string
    message string
}

type ErrorElementWithData struct {
    ErrorElement
    data interface{}
}

func (e ErrorElement) NewErrorWithNewMessage(msg string) ErrorElement {
    return ErrorElement{
        code:    e.code,
        name:    e.name,
        message: msg,
    }
}

func (e ErrorElement) NewErrorWithData(msg string, data interface{}) *ErrorElementWithData {
    return &ErrorElementWithData{
            ErrorElement: ErrorElement{
                code:    e.code,
                name:    e.name,
                message: msg,
            },
            data:    data,
    }
}

func (e ErrorElement)Code() int64  {
    return e.code
}

func (e ErrorElement)Name() string  {
    return e.name
}

func (e ErrorElement)Error() string  {
    return e.message
}

func (e ErrorElementWithData)Data() interface{}  {
   return e.data
}

func BuildErrorEnum(enum interface{}, startCode int64) {
    buildEnum(enum, func(code int, name string, tag reflect.StructTag) reflect.Value {
        codeStr := tag.Get(errorCodeTag)
        codeNum := int64(code) + startCode
        if "" != codeStr {
            codeNum, _ = strconv.ParseInt(codeStr, 0, 64)
        }

        return reflect.ValueOf(ErrorElement{
            code: codeNum,
            name: name,
            message: tag.Get(errorMessageTag),
        })
    })
    return
}
