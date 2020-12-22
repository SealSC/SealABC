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
    "github.com/SealSC/SealABC/network/http"
    "github.com/SealSC/SealABC/service"
    "strconv"
)

type getBlockByHeight struct{
    baseHandler
}

func (g *getBlockByHeight)Handle(ctx *gin.Context) {
    res := http.NewResponse(ctx)
    heightString := ctx.Param(URLParameterKeys.Height.String())

    height, err := strconv.ParseUint(heightString, 10, 64)
    if err != nil {
        res.BadRequest("parameter [height] is not a number")
        return
    }

    blk, err := g.chain.GetBlockRowByHeight(height)
    if err != nil {
        return
    }

    res.OK(blk)
}

func (g *getBlockByHeight)RouteRegister(router gin.IRouter) {
    router.GET(g.buildUrlPath(), g.Handle)
}

func (g *getBlockByHeight)BasicInformation() (info http.HandlerBasicInformation)  {
    info.Description = "return full block data of the given block height."
    info.Path = g.serverBasePath + g.buildUrlPath()
    info.Method = service.ApiProtocolMethod.HttpGet.String()

    info.Parameters.Type = service.ApiParameterType.URL.String()
    info.Parameters.Template = g.serverBasePath + g.urlWithoutParameters() + "/1"
    return
}

func (g *getBlockByHeight) urlWithoutParameters() string {
    return "/get/block/by/height"
}

func (g *getBlockByHeight) buildUrlPath() string {
    return g.urlWithoutParameters() + "/:" + URLParameterKeys.Height.String()
}

