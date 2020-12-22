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

package engineService

import (
    "github.com/SealSC/SealABC/metadata/serviceRequest"
    "github.com/SealSC/SealABC/service"
    "errors"
    "sync"
)

var serviceMap = map[string] service.IService{}
var serviceLock sync.RWMutex

func Mount(s service.IService) (err error) {
    serviceLock.Lock()
    defer serviceLock.Unlock()

    if srv, _ := getService(s.Name()); srv != nil {
        err = errors.New("service named " + s.Name() + " already registered")
        return
    }

    serviceMap[s.Name()] = s
    return
}

func getService(name string) (s service.IService, err error) {
    if _, exists := serviceMap[name]; !exists {
        err = errors.New("service named " + name + " not exists")
        return
    }

    s = serviceMap[name]
    return
}

func preExecuteServiceRequest(req serviceRequest.Entity) (result []byte, err error) {
    serviceLock.Lock()
    defer serviceLock.Unlock()

    s, err := getService(req.RequestService)
    if err != nil {
        return
    }

    return s.PreExecute(req)
}

func executeServiceRequest(req serviceRequest.Entity) (result []byte, err error) {
    serviceLock.Lock()
    defer serviceLock.Unlock()

    s, err := getService(req.RequestService)
    if err != nil {
        return
    }

    return s.Execute(req)
}

func getAllRequestsNeedConsensus() (actions [][]byte) {
    serviceLock.Lock()
    defer serviceLock.Unlock()

    for _, s := range serviceMap {
        act, cnt := s.RequestsForConsensus()
        if cnt == 0 {
            continue
        }
        actions = append(actions, act...)
    }

    return
}
