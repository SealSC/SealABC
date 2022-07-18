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

package tsLedger

import (
	"errors"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/crypto/signers/ed25519"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/service/application/traceableStorage/tsData"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

type tsServiceValidator func(data tsData.TSData) (err error)
type tsServiceActuator func(data tsData.TSData) (ret interface{}, err error)
type tsQuery func(param []string) (ret interface{}, err error)

type TSLedger struct {
	reqPool []tsData.TSServiceRequest

	validators map[string]tsServiceValidator
	actuators  map[string]tsServiceActuator

	CryptoTools crypto.Tools
	Storage     kvDatabase.IDriver
}

func Load() {
	enum.SimpleBuild(&tsData.RequestTypes)
}

func (t *TSLedger) VerifyRequest(req tsData.TSServiceRequest) (err error) {
	if validator, exists := t.validators[req.ReqType]; exists {
		return validator(req.Data)
	} else {
		return errors.New("invalid request type")
	}
}

func (t *TSLedger) ExecuteRequest(req tsData.TSServiceRequest) (ret interface{}, err error) {
	if actuator, exists := t.actuators[req.ReqType]; exists {
		return actuator(req.Data)
	} else {
		return nil, errors.New("invalid actuator type")
	}
}

func NewTraceableStorage(kvDriver kvDatabase.IDriver, sqlDriver simpleSQLDatabase.IDriver) *TSLedger {
	t := &TSLedger{
		CryptoTools: crypto.Tools{
			HashCalculator:  sha3.Sha256,
			SignerGenerator: ed25519.SignerGenerator,
		},
		Storage: kvDriver,
	}

	t.validators = map[string]tsServiceValidator{
		tsData.RequestTypes.Store.String():  t.VerifyStoreRequest,
		tsData.RequestTypes.Modify.String(): t.VerifyModifyRequest,
	}

	t.actuators = map[string]tsServiceActuator{
		tsData.RequestTypes.Store.String():  t.ExecuteStoreIdentification,
		tsData.RequestTypes.Modify.String(): t.ExecuteModifyIdentification,
	}

	return t
}
