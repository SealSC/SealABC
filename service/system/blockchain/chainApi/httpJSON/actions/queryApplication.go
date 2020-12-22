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
    "github.com/SealSC/SealABC/metadata/blockchainRequest"
    "github.com/SealSC/SealABC/network/http"
    "github.com/SealSC/SealABC/service"
    "github.com/gin-gonic/gin"
)

type queryApplication struct{
    baseHandler
}

func (q *queryApplication)Handle(ctx *gin.Context) {
    res := http.NewResponse(ctx)

    appName := ctx.Param(URLParameterKeys.App.String())

    handler, exist := q.appQueryHandler[appName]
    if !exist {
        res.BadRequest("no such application: " + appName)
        return
    }

    reqData, err := ctx.GetRawData()
    if err != nil {
        res.BadRequest(err.Error())
        return
    }

    ret, err := handler(reqData)
    if err != nil {
        res.BadRequest(err.Error())
        return
    }

    res.OK(ret)
}

func (q *queryApplication)RouteRegister(router gin.IRouter) {
    router.POST(q.buildUrlPath(), q.Handle)
}

func (q *queryApplication)BasicInformation() (info http.HandlerBasicInformation) {

    info.Description = "will execute application query operation that registered on the blockchain."
    info.Path = q.serverBasePath + q.buildUrlPath()
    info.Method = service.ApiProtocolMethod.HttpPost.String()

    info.Parameters.Type = service.ApiParameterType.JSON.String()
    info.Parameters.Template = blockchainRequest.Entity{}
    return
}

func (q *queryApplication) urlWithoutParameters() string  {
    return "/query/application"
}

func (q *queryApplication) buildUrlPath() string {
    return q.urlWithoutParameters() + "/:" + URLParameterKeys.App.String()
}

