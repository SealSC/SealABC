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

package oracleInterface

import (
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"crypto/tls"
	"net/http"
)

func isBaseAction(a Action) bool {
	switch a.(type) {
	case ActionRemoteVerifier,
		ActionRemoteAutoPuller,
		ActionRequestSaver,
		ActionRemoteManualPuller:
		return false
	}
	return true
}

type Action interface {
	Name() string
	FormatResult(req blockchainRequest.Entity) (interface{}, error)
}

type ActionRemoteVerifier interface {
	Action
	RemoteGet
	RequestVerifier
	EqualSAB()
}

type ActionRemoteManualPuller interface {
	Action
	RemoteGet
	RequestVerifier
}

type ActionRemoteAutoPuller interface {
	RequestVerifier
	Action
	AutoFunc
	RemoteGet
}

type ActionRequestSaver interface {
	Action
	RequestVerifier
}

type AutoFunc interface {
	//CronPaths data format{Second | Minute | Hour | Dom | Month | Dow | Descriptor}
	CronPaths() []string
}

type RequestVerifier interface {
	VerifyReq(r blockchainRequest.Entity) error
}
type RemoteGet interface {
	VerifyRemoteResp(*http.Response) (blockchainRequest.Entity, error)
	VerifyRemoteCA(state tls.ConnectionState) error
	UrlContentType() (string, string)
}
