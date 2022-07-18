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

package memoSQLStorage

import (
	"encoding/hex"
	"errors"
	"strconv"
)

func pageFromParam(param []string) (page uint64, err error) {
	if len(param) != 1 {
		err = errors.New("invalid parameters")
		return
	}

	page, err = strconv.ParseUint(param[0], 10, 64)
	return
}

func hashFromParam(param []string) (hash string, err error) {
	if len(param) != 1 {
		err = errors.New("invalid parameters")
		return
	}

	_, err = hex.DecodeString(param[0])
	if err != nil {
		return
	}

	hash = param[0]
	return
}

func pageAndHashFromParam(param []string) (page uint64, hash string, err error) {
	if len(param) != 2 {
		err = errors.New("invalid parameters")
		return
	}

	page, err = strconv.ParseUint(param[0], 10, 64)
	if err != nil {
		return
	}

	_, err = hex.DecodeString(param[1])
	if err != nil {
		return
	}

	hash = param[1]
	return
}

func pageAndMemoTypeFromParam(param []string) (page uint64, memoType string, err error) {
	if len(param) != 2 {
		err = errors.New("invalid parameters")
		return
	}

	page, err = strconv.ParseUint(param[0], 10, 64)
	if err != nil {
		return
	}

	memoType = param[1]
	return
}
