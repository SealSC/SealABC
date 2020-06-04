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
)

func buildEnum(enum interface{}, isErrEnum bool, valueBuilder func(int, string, reflect.StructTag) reflect.Value) {
    eValue := reflect.ValueOf(enum).Elem()
    eType := eValue.Type()

    for i:=0; i<eValue.NumField(); i++ {
        elemValue := eValue.Field(i)

        if !elemValue.CanSet() {
            continue
        }

        if !elemValue.CanInterface() {
            continue
        }

        if isErrEnum {
            _, ok := elemValue.Interface().(ErrorElement)
            if !ok {
                continue
            }
        } else {
            _, ok := elemValue.Interface().(Element)
            if !ok {
                continue
            }
        }

        elemType := eType.Field(i)
        enumValue := valueBuilder(i, elemType.Name, elemType.Tag)
        elemValue.Set(enumValue)
    }

    return
}
