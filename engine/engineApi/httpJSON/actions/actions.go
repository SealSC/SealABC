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

package actions

import (
	"github.com/SealSC/SealABC/network/http"
	"github.com/SealSC/SealABC/service"
)

var actionList []http.IRequestHandler
var serverConfig http.Config

func Load(cfg http.Config) []http.IRequestHandler {
	serverConfig = cfg

	actionList = []http.IRequestHandler{
		ListServices,
	}

	return actionList
}

func ApiInformation() (list []service.ApiInterface) {

	for _, act := range actionList {
		actInfo := act.BasicInformation()
		ai := service.ApiInterface{}

		ai.Description = actInfo.Description
		ai.Path = actInfo.Path
		ai.Method = actInfo.Method
		ai.Parameters = (service.Parameters)(actInfo.Parameters)

		list = append(list, ai)
	}

	return
}
