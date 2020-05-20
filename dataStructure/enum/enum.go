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

import "reflect"

const (
    elementDescriptionTag = "desc"
    elementNameTag = "name"
)

type IElement interface {
    Int() int
    String() string
    Description() string
}

type Element struct {
    value       int
    name        string
    description string
}

func (t *Element)Int() int  {
    return t.value
}

func (t *Element)String() string  {
    return t.name
}

func (t *Element)Description() string {
    return t.name
}

func Build(enums interface{}, begin int, prefix string) {
    buildEnum(enums, func(value int, name string, tag reflect.StructTag) reflect.Value {
        taggedName := tag.Get(elementNameTag)
        enumName := prefix + name
        if taggedName != "" {
            enumName = taggedName
        }

        return reflect.ValueOf(Element{
            value: value + begin,
            name: enumName,
            description: tag.Get(elementDescriptionTag),
        })
    })
    return
}

func SimpleBuild(enums interface{}) {
    Build(enums, 0, "")
}
