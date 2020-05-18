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
    "github.com/gin-gonic/gin"
    "SealABC/engine/engineService"
    "SealABC/network/http"
    "SealABC/service"
)

type listServices struct{
    path string
}

var ListServices = &listServices{
    path: "/list/services",
}

type engineSetting struct {
    Consensus interface{}
}

func (c *listServices)Handle(ctx *gin.Context) {
    res := http.NewResponse(ctx)

    basicInfo := service.BasicInformation{}

    es := engineSetting{}

    basicInfo.Name = "Seal ABC"
    basicInfo.Description = "the engine of Seal ABC blockchain framework"
    settingsInfo, subService := engineService.GetServicesBasicInformation()
    es.Consensus = settingsInfo
    basicInfo.Settings = es
    basicInfo.SubServices = subService
    basicInfo.Api.Protocol = service.ApiProtocols.HTTP.String()
    basicInfo.Api.Address = serverConfig.Address
    basicInfo.Api.ApiList = ApiInformation()

    res.OK(basicInfo)
}

func (c *listServices)RouteRegister(router gin.IRouter) {
    router.GET(serverConfig.BasePath + c.path, c.Handle)
}

func (c *listServices)BasicInformation() (info http.HandlerBasicInformation)  {

    info.Description = "this method will list all the services mounted on this engine."
    info.Path = serverConfig.BasePath + c.path
    info.Method = service.ApiProtocolMethod.HttpGet.String()

    return
}

