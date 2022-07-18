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
	"errors"
	"github.com/SealSC/SealABC/log"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	Config *Config
}

func (s *Server) Start() (err error) {
	if s.Config == nil {
		err = errors.New("no configure")
		return
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		if s.Config.AllowCORS {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		if c.Request.Method == "OPTIONS" {
			if len(c.Request.Header["Access-Control-Request-Headers"]) > 0 {
				c.Header("Access-Control-Allow-Headers", c.Request.Header["Access-Control-Request-Headers"][0])
			}
			c.AbortWithStatus(http.StatusNoContent)
		} else {
			c.Next()
		}
	})
	s.setRouters(router, *s.Config)

	go func() {
		runSrvErr := router.Run(s.Config.Address)
		if runSrvErr != nil {
			log.Log.Warn("start http server failed: ", runSrvErr.Error())
		}
	}()

	return
}

func (s *Server) setRouters(router *gin.Engine, cfg Config) {
	for _, v := range cfg.RequestHandler {
		v.RouteRegister(router)
	}
}
