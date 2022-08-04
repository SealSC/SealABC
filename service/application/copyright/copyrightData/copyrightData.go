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

package copyrightData

import (
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/metadata/seal"
)

type CopyrightResource struct {
	ID           []byte
	Owner        []byte
	ResourceHash []byte
}

type CopyrightCertificate struct {
	ID            []byte
	Applicant     []byte
	ApplicantSeal seal.Entity

	Owner   []byte
	OwnSeal seal.Entity

	BasicHash []byte
	ExtraHash []byte

	CustomData []byte

	PlatformSeal seal.Entity

	IssuanceTransaction []byte
	IssuanceTime        string
	IssuanceType        string
	Algorithm           crypto.AlgorithmDescription
}

type CopyrightTrading struct {
}

const APPName = "Traceable Storage"
