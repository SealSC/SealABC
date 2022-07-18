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

package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ServiceResult struct {
	Code    int64       `json:"code"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type Response struct {
	ctx *gin.Context
}

func NewResponse(ctx *gin.Context) *Response {
	return &Response{
		ctx: ctx,
	}
}

func (c *Response) BadRequest(data interface{}) {
	c.ctx.JSON(http.StatusBadRequest, data)
}

func (c *Response) ForbiddenRequest(data interface{}) {
	c.ctx.JSON(http.StatusForbidden, data)
}

func (c *Response) NotFoundRequest(data interface{}) {
	c.ctx.JSON(http.StatusNotFound, data)
}

func (c *Response) InternalServerError(data interface{}) {
	c.ctx.JSON(http.StatusInternalServerError, data)
}

func (c *Response) OK(data interface{}) {
	c.ctx.JSON(http.StatusOK, data)
}

//service response will always set http status to 200.
func (c *Response) ServiceError(code int64, data interface{}) {
	c.ctx.JSON(http.StatusOK, &ServiceResult{
		code, false, data,
	})
}

func (c *Response) ServiceSuccess(data interface{}) {
	c.ctx.JSON(http.StatusOK, &ServiceResult{
		0, true, data,
	})
}
