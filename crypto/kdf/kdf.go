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

package kdf

type IKeyDerivationFunction interface {
	Name() string
	NewKey(password []byte, keyLen int) (key []byte, salt []byte, param []byte, err error)
	NewKeyWithParam(password []byte, keyLen int, param []byte) (key []byte, salt []byte, err error)
	RebuildKey(password []byte, keyLen int, salt []byte, param []byte) (key []byte, err error)
}
