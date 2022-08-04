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

type getBlockByHash struct {
	baseHandler
}

func (g *getBlockByHash) Handle(ctx *gin.Context) {
	res := http.NewResponse(ctx)
	hash := ctx.Param(URLParameterKeys.HexHash.String())
	blk, err := g.chain.GetBlockRowByHash(hash)
	if err != nil {
		res.BadRequest(err.Error())
		return
	}

	res.OK(blk)
}

func (g *getBlockByHash) RouteRegister(router gin.IRouter) {
	router.GET(g.buildUrlPath(), g.Handle)
}

func (g *getBlockByHash) BasicInformation() (info http.HandlerBasicInformation) {
	info.Description = "return full block data of the given block hash."
	info.Path = g.serverBasePath + g.buildUrlPath()
	info.Method = service.ApiProtocolMethod.HttpGet.String()

	info.Parameters.Type = service.ApiParameterType.URL.String()
	info.Parameters.Template = g.serverBasePath + g.urlWithoutParameters() + "/1ae9d62bea40f591af7ab6e03e077d85adb33a66cd977e913763a303599c5440"
	return
}

func (g *getBlockByHash) urlWithoutParameters() string {
	return "/get/block/by/hash"
}

func (g *getBlockByHash) buildUrlPath() string {
	return g.urlWithoutParameters() + "/:" + URLParameterKeys.HexHash.String()
}
