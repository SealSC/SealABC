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
    "strconv"
)

type getTransactions struct{
    baseHandler
}

func (g *getTransactions)Handle(ctx *gin.Context) {
    res := http.NewResponse(ctx)
    pageString := ctx.Param(URLParameterKeys.Page.String())

    page, err := strconv.ParseUint(pageString, 10, 64)
    if err != nil {
        res.BadRequest("parameter [page] is not a number")
        return
    }

    ret, err := g.sqlStorage.GetRequestList(page)
    if err != nil {
        res.InternalServerError(err.Error())
        return
    }

    res.OK(ret)
}

func (g *getTransactions)RouteRegister(router gin.IRouter) {
    router.GET(g.buildUrlPath(), g.Handle)
}

func (g *getTransactions)BasicInformation() (info http.HandlerBasicInformation)  {
    info.Description = "return block data list."
    info.Path = g.serverBasePath + g.buildUrlPath()
    info.Method = service.ApiProtocolMethod.HttpGet.String()

    info.Parameters.Type = service.ApiParameterType.URL.String()
    info.Parameters.Template = g.serverBasePath + g.urlWithoutParameters() + "/1"
    return
}

func (g *getTransactions) urlWithoutParameters() string {
    return "/get/request/list/page/"
}

func (g *getTransactions) buildUrlPath() string {
    return g.urlWithoutParameters() + "/:" + URLParameterKeys.Page.String()
}


