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
	"SealABC/network/http"
	"SealABC/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type getTransactionByApplicationAndAction struct{
	baseHandler
}

func (g *getTransactionByApplicationAndAction)Handle(ctx *gin.Context) {
	res := http.NewResponse(ctx)
	app := ctx.Param(URLParameterKeys.App.String())
	act := ctx.Param(URLParameterKeys.Action.String())
	page := ctx.Param(URLParameterKeys.Page.String())

	pageNum, err := strconv.ParseUint(page, 10, 64)
	if err != nil {
		res.BadRequest("invalid page")
		return
	}

	txList, err := g.sqlStorage.GetRequestByApplicationAndAction(app, act, pageNum)
	if err != nil {
		res.InternalServerError(err.Error())
		return
	}

	res.OK(txList)
}

func (g *getTransactionByApplicationAndAction)RouteRegister(router gin.IRouter) {
	router.GET(g.buildUrlPath(), g.Handle)
}

func (g *getTransactionByApplicationAndAction)BasicInformation() (info http.HandlerBasicInformation)  {
	info.Description = "return full block data of the given block hash."
	info.Path = g.serverBasePath + g.buildUrlPath()
	info.Method = service.ApiProtocolMethod.HttpGet.String()

	info.Parameters.Type = service.ApiParameterType.URL.String()
	info.Parameters.Template = g.serverBasePath + g.urlWithoutParameters() + "/1"
	return
}

func (g *getTransactionByApplicationAndAction) urlWithoutParameters() string  {
	return "/get/request/by/application/"
}


func (g *getTransactionByApplicationAndAction) buildUrlPath() string {
	return g.urlWithoutParameters() +
		"/:" + URLParameterKeys.App.String() +
		"/:" + URLParameterKeys.Action.String() +
		"/:" + URLParameterKeys.Page.String()
}

