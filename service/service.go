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

package service

import (
    "errors"
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/network/http"
)

type Parameters http.Parameters

var ApiProtocols = struct {
    HTTP        enum.Element
    TCP         enum.Element
    UDP         enum.Element
    INTERNAL    enum.Element
}{}

var ApiProtocolMethod = struct {
    HttpGet      enum.Element
    HttpPost     enum.Element
    Binary       enum.Element
    InternalCall enum.Element
}{}

var ApiParameterType = struct {
    XML     enum.Element
    JSON    enum.Element
    URL     enum.Element
    Binary  enum.Element
    Private enum.Element
}{}

type ApiInterface struct {
    Description string
    Path        string
    Method      string
    Parameters  Parameters
}

type ApiDescription struct {
    Protocol    string
    Address     string
    ApiList     []ApiInterface
}

type BasicInformation struct {
    Name        string
    Description string
    Settings    interface {}
    Api         ApiDescription
    SubServices []BasicInformation
}


type IService interface {
    //service name, this is the ID that registered to engine
    Name() (name string)

    //build request list for consensus network, engine will call this method when you are a consensus leader
    RequestsForConsensus() (req [][]byte, cnt uint32)

    //receive and verify request from consensus network (on the other word, this request comes from consensus leader)
    PreExecute(req interface{}) (result []byte, err error)

    //receive and execute confirmed request of this round of consensus
    Execute(req interface{}) (result []byte, err error)

    //handle consensus failed
    Cancel(req interface{}) (err error)

    //service information
    Information() (info BasicInformation)
}

//a blank service to simplify the creation of new services
type BlankService struct {}

func (b BlankService) Name() (name string) {
    return
}

func (b BlankService) PushClientRequest(req interface{}) (result string, err error) {
    err = errors.New("not support")
    return
}

func (b BlankService) RequestsForConsensus() (req [][]byte, cnt uint32) {
    return
}

func (b BlankService) PreExecute(req interface{}) (result []byte, err error) {
    return
}

func (b BlankService) Execute(req interface{}) (result []byte, err error) {
    return
}

func (b BlankService) Cancel(req interface{}) (err error) {
    return
}

func (b BlankService) Information() (info BasicInformation) {
    return
}

func Load() {
    enum.SimpleBuild(&ApiProtocols)
    enum.SimpleBuild(&ApiProtocolMethod)
    enum.SimpleBuild(&ApiParameterType)
}
