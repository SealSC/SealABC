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
	"github.com/SealSC/SealABC/dataStructure/mfb"
	"reflect"
)

type standardSerializer func(value reflect.Value) ([]byte, error)

var standardSerializerMap map[reflect.Kind]standardSerializer

func loadSerializer() {
	standardSerializerMap = map[reflect.Kind]standardSerializer{
		reflect.Bool: boolToBytes,

		reflect.Int:   intToBytes,
		reflect.Int8:  intToBytes,
		reflect.Int16: intToBytes,
		reflect.Int32: intToBytes,
		reflect.Int64: intToBytes,

		reflect.Uint:   uintToBytes,
		reflect.Uint8:  uintToBytes,
		reflect.Uint16: uintToBytes,
		reflect.Uint32: uintToBytes,
		reflect.Uint64: uintToBytes,

		reflect.String: stringToBytes,

		reflect.Array: arrayToBytes,
		reflect.Slice: arrayToBytes,

		reflect.Interface: interfaceToBytes,
		reflect.Ptr:       interfaceToBytes,
		reflect.Struct:    structToBytes,
	}
}

func callStandardSerializer(kind reflect.Kind, value reflect.Value) (data []byte, err error) {
	serializer, exist := standardSerializerMap[kind]
	if !exist {
		return
	}

	data, err = serializer(value)

	return
}

func boolToBytes(value reflect.Value) (data []byte, err error) {
	if value.Bool() {
		data = []byte{1}
	} else {
		data = []byte{0}
	}
	return
}

func intToBytes(value reflect.Value) (data []byte, err error) {
	return commonIntToBytes(uint64(value.Int()), value.Kind())
}

func uintToBytes(value reflect.Value) (data []byte, err error) {
	return commonIntToBytes(value.Uint(), value.Kind())
}

func commonIntToBytes(i uint64, kind reflect.Kind) (data []byte, err error) {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		data = make([]byte, 1, 1)
		data[0] = byte(i)

	case reflect.Int16, reflect.Uint16:
		data = make([]byte, 2, 2)
		binary.BigEndian.PutUint16(data, uint16(i))

	case reflect.Int32, reflect.Uint32:
		data = make([]byte, 4, 4)
		binary.BigEndian.PutUint32(data, uint32(i))

	case reflect.Int64, reflect.Int, reflect.Uint, reflect.Uint64:
		data = make([]byte, 8, 8)
		binary.BigEndian.PutUint64(data, i)
	}

	return
}

func stringToBytes(value reflect.Value) (data []byte, err error) {
	data = []byte(value.String())
	return
}

func arrayToBytes(value reflect.Value) (data []byte, err error) {
	sLen := value.Len()

	sKind := value.Type().Elem().Kind()
	var sElBytes []byte

	if sKind == reflect.Struct {
		for i := 0; i < sLen; i += 1 {
			sElBytes, err = toMFBytes(value.Index(i), true)
			if err != nil {
				return
			}

			var mfBytes mfb.MarkedFlatBytes
			mfBytes.FromByteSlice([][]byte{sElBytes})
			data = append(data, mfBytes...)
		}
	} else if isMarklessSlice(sKind) {
		if sKind == reflect.Uint8 {
			data = make([]byte, sLen, sLen)
			copy(data, value.Bytes())
		} else {
			for i := 0; i < sLen; i += 1 {
				sElBytes, err = callStandardSerializer(sKind, value.Index(i))
				if err != nil {
					return
				}

				data = append(data, sElBytes...)
			}
		}
	} else {
		for i := 0; i < sLen; i += 1 {
			sElBytes, err = callStandardSerializer(sKind, value.Index(i))
			if err != nil {
				return
			}

			var mfBytes mfb.MarkedFlatBytes
			mfBytes.FromByteSlice([][]byte{sElBytes})
			data = append(data, mfBytes...)
		}
	}
	return
}

func interfaceToBytes(value reflect.Value) (data []byte, err error) {
	if !value.CanInterface() {
		return
	}

	el := value.Elem()

	if el.Kind() == reflect.Struct {
		return toMFBytes(el, true)
	} else {
		return callStandardSerializer(el.Kind(), el)
	}

	return
}

func structToBytes(value reflect.Value) (data []byte, err error) {
	if !value.CanInterface() {
		return
	}

	return toMFBytes(value.Interface(), false)
}

func toMFBytes(st interface{}, reflected bool) (data mfb.MarkedFlatBytes, err error) {
	stType := reflect.TypeOf(st)

	var stElements reflect.Value
	if reflected {
		stElements = st.(reflect.Value)
	} else {
		if stType.Kind() != reflect.Struct {
			return
		}

		stElements = reflect.ValueOf(st)
	}

	var bytesList [][]byte

	for i := 0; i < stElements.NumField(); i++ {
		elValue := stElements.Field(i)
		if !elValue.CanInterface() {
			continue
		}

		//if elValue.Kind() == reflect.Interface {
		//    continue
		//}

		elKind := elValue.Kind()

		var elBytes []byte

		elBytes, err = callStandardSerializer(elKind, elValue)
		if err != nil {
			break
		}

		bytesList = append(bytesList, elBytes)
	}

	data.FromByteSlice(bytesList)
	return
}

func ToMFBytes(st interface{}) (data mfb.MarkedFlatBytes, err error) {
	return toMFBytes(st, false)
}
