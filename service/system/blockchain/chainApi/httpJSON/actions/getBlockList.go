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
    "github.com/SealSC/SealABC/service/system/blockchain/chainTables"
    "errors"
    "github.com/gin-gonic/gin"
    "strconv"
)

type getBlockList struct{
    baseHandler
}

type getBlockListResult struct {
    Blocks          []chainTables.BlockListRow
    CurrentHeight   uint64
}

func (g *getBlockList)getStartHeight(pageString string, countInPage uint64) (startHeight uint64, currentHeight uint64, err error) {
    currentHeight = g.chain.CurrentHeight()
    if pageString == "last" {
        startHeight = 0
        return
    }

    page, err := strconv.ParseUint(pageString, 10, 64)
    if err != nil {
        err = errors.New("parameter [page] is not valid")
        return
    }

    heightOffset := ((page + 1) * countInPage) - 1
    if heightOffset >= currentHeight {
        startHeight = 0
        return
    }

    startHeight = currentHeight - heightOffset
    return
}

func (g *getBlockList)Handle(ctx *gin.Context) {
    res := http.NewResponse(ctx)
    pageString := ctx.Param(URLParameterKeys.Page.String())

    const countInPage = 20

    startHeight, currentHeight, err := g.getStartHeight(pageString, countInPage)
    if err != nil {
        res.BadRequest(err.Error())
        return
    }

    blkList, err := g.sqlStorage.GetBlockList(startHeight)
    if err != nil {
        res.InternalServerError(err.Error())
        return
    }

    res.OK(getBlockListResult{
        Blocks:        blkList,
        CurrentHeight: currentHeight,
    })
}

func (g *getBlockList)RouteRegister(router gin.IRouter) {
    router.GET(g.buildUrlPath(), g.Handle)
}

func (g *getBlockList)BasicInformation() (info http.HandlerBasicInformation)  {
    info.Description = "return block data list."
    info.Path = g.serverBasePath + g.buildUrlPath()
    info.Method = service.ApiProtocolMethod.HttpGet.String()

    info.Parameters.Type = service.ApiParameterType.URL.String()
    info.Parameters.Template = g.serverBasePath + g.urlWithoutParameters() + "/1"
    return
}

func (g *getBlockList) urlWithoutParameters() string {
    return "/get/block/list/page/"
}

func (g *getBlockList) buildUrlPath() string {
    return g.urlWithoutParameters() + "/:" + URLParameterKeys.Page.String()
}
