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

package structSerializer

import (
	"encoding/binary"
	"errors"
	"github.com/SealSC/SealABC/dataStructure/mfb"
	"reflect"
)

type standardDeserializer func(value reflect.Value, data []byte) error

var standardDeserializerMap map[reflect.Kind]standardDeserializer

func loadDeserializer() {
	standardDeserializerMap = map[reflect.Kind]standardDeserializer{
		reflect.Bool: bytesToBool,

		reflect.Int:   bytesToInt,
		reflect.Int8:  bytesToInt,
		reflect.Int16: bytesToInt,
		reflect.Int32: bytesToInt,
		reflect.Int64: bytesToInt,

		reflect.Uint:   bytesToUint,
		reflect.Uint8:  bytesToUint,
		reflect.Uint16: bytesToUint,
		reflect.Uint32: bytesToUint,
		reflect.Uint64: bytesToUint,

		reflect.String: bytesToString,

		reflect.Array: bytesToArray,
		reflect.Slice: bytesToArray,

		reflect.Interface: bytesToInterface,
		reflect.Ptr:       bytesToInterface,
		reflect.Struct:    bytesToStruct,
	}
}

func getMarklessElementLen(kind reflect.Kind) (len int) {
	switch kind {
	case reflect.Bool:
		len = 1
	case reflect.Int8, reflect.Uint8:
		len = 1
	case reflect.Int16, reflect.Uint16:
		len = 2
	case reflect.Int32, reflect.Uint32:
		len = 4
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		len = 8
	}

	return
}

func bytesToBool(value reflect.Value, data []byte) (err error) {
	if 1 == data[0] {
		value.SetBool(true)
	} else {
		value.SetBool(false)
	}
	return
}

func bytesToInt(value reflect.Value, data []byte) (err error) {
	var dataVal int64
	switch value.Kind() {
	case reflect.Int8:
		dataVal = int64(data[0])
	case reflect.Int16:
		dataVal = int64(binary.BigEndian.Uint16(data))
	case reflect.Int32:
		dataVal = int64(binary.BigEndian.Uint32(data))
	case reflect.Int, reflect.Int64:
		dataVal = int64(binary.BigEndian.Uint64(data))
	}
	value.SetInt(dataVal)
	return
}

func bytesToUint(value reflect.Value, data []byte) (err error) {
	var dataVal uint64
	switch value.Kind() {
	case reflect.Uint8:
		dataVal = uint64(data[0])
	case reflect.Uint16:
		dataVal = uint64(binary.BigEndian.Uint16(data))
	case reflect.Uint32:
		dataVal = uint64(binary.BigEndian.Uint32(data))
	case reflect.Uint, reflect.Uint64:
		dataVal = uint64(binary.BigEndian.Uint64(data))
	}
	value.SetUint(dataVal)
	return
}

func bytesToString(value reflect.Value, data []byte) (err error) {
	value.SetString(string(data))
	return
}

type bytesToArrayMethod func(el reflect.Value, data []byte, kind reflect.Kind) (err error)

func bytesToComplexArray(el reflect.Value, data []byte, kind reflect.Kind) (err error) {
	return callStandardDeserializer(kind, el, data)
}

func bytesToStructArray(el reflect.Value, data []byte, kind reflect.Kind) (err error) {
	return fromMFBytes(data, el, true)
}

func bytesToArray(value reflect.Value, data []byte) (err error) {
	elKind := value.Type().Elem().Kind()
	sType := value.Type()

	var arrayFromReflect reflect.Value
	var sMakeType reflect.Type
	if sType.Kind() == reflect.Slice {
		sMakeType = sType
	} else {
		sMakeType = reflect.SliceOf(value.Type().Elem())
	}

	sElLen := getMarklessElementLen(elKind)
	if !isMarklessSlice(elKind) {
		mfBytes := mfb.MarkedFlatBytes(data)
		var bytesList [][]byte
		bytesList, err = mfBytes.ToByteSlice()
		if err != nil {
			return
		}

		sLen := len(bytesList)

		if len(bytesList) == 0 {
			return
		}

		arrayFromReflect = reflect.MakeSlice(sMakeType, sLen, sLen)

		var deserializeMethod bytesToArrayMethod
		if elKind == reflect.Struct {
			deserializeMethod = bytesToStructArray
		} else {
			deserializeMethod = bytesToComplexArray
		}

		for i := 0; i < sLen; i += 1 {
			err = deserializeMethod(arrayFromReflect.Index(i), bytesList[i], elKind)
			if err != nil {
				return
			}
		}
	} else {
		sLen := len(data) / sElLen
		arrayFromReflect = reflect.MakeSlice(sMakeType, sLen, sLen)

		if sElLen != 1 {
			for i := 0; i < sLen; i += 1 {
				err = callStandardDeserializer(elKind, arrayFromReflect.Index(i), data[i:(i+1)*sElLen])
				if err != nil {
					return
				}
			}
		}
	}

	if sType.Kind() == reflect.Slice {
		if sElLen == 1 {
			value.SetBytes(data)
		} else {
			value.Set(arrayFromReflect)
		}
	} else {
		reflect.Copy(value, arrayFromReflect)
	}
	return
}

func bytesToInterface(value reflect.Value, data []byte) (err error) {
	if !value.CanInterface() || len(data) == 0 {
		return
	}

	pointToElType := reflect.TypeOf(value.Interface()).Elem()

	el := reflect.New(pointToElType)

	if pointToElType.Kind() == reflect.Struct {
		var mfBytes mfb.MarkedFlatBytes = data
		bytesList, _ := mfBytes.ToByteSlice()
		if len(bytesList) == 0 {
			println()
			return
		}
		fromMFBytes(data, el.Elem(), true)
	} else {
		callStandardDeserializer(pointToElType.Kind(), el.Elem(), data)
	}

	value.Set(el)

	return
}

func bytesToStruct(value reflect.Value, data []byte) (err error) {
	return fromMFBytes(data, value, true)
}

func callStandardDeserializer(kind reflect.Kind, value reflect.Value, data []byte) (err error) {
	deserializer, exist := standardDeserializerMap[kind]
	if !exist {
		return
	}

	err = deserializer(value, data)

	return
}

func fromMFBytes(
	mfBytes mfb.MarkedFlatBytes,
	toST interface{},
	reflected bool) (err error) {

	stType := reflect.TypeOf(toST)
	if stType == nil {
		return errors.New("can't detect type of input.")
	}

	var stElements reflect.Value
	if reflected {
		stElements = toST.(reflect.Value)
	} else {
		if stType.Kind() != reflect.Ptr {
			return
		}

		stElements = reflect.ValueOf(toST).Elem()
	}

	bytesList, err := mfBytes.ToByteSlice()
	if err != nil {
		return
	}

	if len(bytesList) == 0 {
		return
	}

	j := 0
	for i := 0; i < stElements.NumField(); i++ {
		elValue := stElements.Field(i)
		if !elValue.CanInterface() {
			continue
		}

		//if elValue.Kind() == reflect.Interface {
		//    continue
		//}

		j += 1
		if !elValue.CanSet() {
			continue
		}

		elKind := elValue.Kind()

		err = callStandardDeserializer(elKind, elValue, bytesList[j-1])
		if err != nil {
			break
		}
	}

	return
}

func FromMFBytes(mfBytes mfb.MarkedFlatBytes, toST interface{}) (err error) {
	return fromMFBytes(mfBytes, toST, false)
}
