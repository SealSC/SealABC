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
	"github.com/gin-gonic/gin"
)

type getTransactionByHeight struct {
	baseHandler
}

func (g *getTransactionByHeight) Handle(ctx *gin.Context) {
	res := http.NewResponse(ctx)
	height := ctx.Param(URLParameterKeys.Height.String())

	txList, err := g.sqlStorage.GetRequestByHeight(height)
	if err != nil {
		res.InternalServerError(nil)
		return
	}

	res.OK(txList)
}

func (g *getTransactionByHeight) RouteRegister(router gin.IRouter) {
	router.GET(g.buildUrlPath(), g.Handle)
}

func (g *getTransactionByHeight) BasicInformation() (info http.HandlerBasicInformation) {
	info.Description = "return requests list of the given block height."
	info.Path = g.serverBasePath + g.buildUrlPath()
	info.Method = service.ApiProtocolMethod.HttpGet.String()

	info.Parameters.Type = service.ApiParameterType.URL.String()
	info.Parameters.Template = g.serverBasePath + g.urlWithoutParameters() + "/1"
	return
}

func (g *getTransactionByHeight) urlWithoutParameters() string {
	return "/get/request/by/height"
}

func (g *getTransactionByHeight) buildUrlPath() string {
	return g.urlWithoutParameters() + "/:" + URLParameterKeys.Height.String()
}
