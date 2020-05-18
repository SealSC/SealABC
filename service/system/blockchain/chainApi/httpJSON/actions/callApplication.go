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
    "SealABC/log"
    "SealABC/metadata/blockchainRequest"
    "SealABC/network/http"
    "SealABC/service"
    "github.com/gin-gonic/gin"
)

type callApplication struct{
    baseHandler
}

func (c *callApplication)Handle(ctx *gin.Context) {
    res := http.NewResponse(ctx)

    reqData := blockchainRequest.Entity{}
    _, err := http.GetPostedJson(ctx, &reqData)
    if err != nil {
        res.BadRequest(err.Error())
        return
    }

    result, err := c.chain.Executor.PushRequest(reqData)

    if err != nil {
        res.BadRequest(err.Error())
        return
    }

    broadcastErr := c.p2p.BroadcastRequest(reqData)
    if broadcastErr != nil{
        log.Log.Warn("broadcast request failed: ", broadcastErr)
    }
    res.OK(result)
}

func (c *callApplication)   RouteRegister(router gin.IRouter) {
    router.POST(c.buildUrlPath(), c.Handle)
}

func (c *callApplication)BasicInformation() (info http.HandlerBasicInformation) {

    info.Description = "will call application that registered on the blockchain."
    info.Path = c.serverBasePath + c.buildUrlPath()
    info.Method = service.ApiProtocolMethod.HttpPost.String()

    info.Parameters.Type = service.ApiParameterType.JSON.String()
    info.Parameters.Template = blockchainRequest.Entity{}
    return
}

func (c *callApplication) buildUrlPath() string {
    return "/call/application"
}
