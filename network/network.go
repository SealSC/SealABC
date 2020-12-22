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

package network

import (
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/message"
)

type StaticInformation struct {
    Topology        string
    ConnectedNode   []string
}

type IService interface {
    Self() (node Node)

    Create(cfg Config) (err error)
    ConnectTo(Node) (err error)

    Join(seeds []Node, cfg *Config) (err error)
    Leave()

    GetAllLinkedNode() (nodes []Node)

    SendTo(node Node, msg message.Message) (n int, err error)
    Broadcast(msg message.Message) (err error)

    RegisterMessageProcessor(msgFamily string, processor MessageProcessor)

    StaticInformation() StaticInformation
}

type Service struct {
    started bool
    router  IRouter
}

func NewNetworkNodeFromLink(link ILink) (node LinkNode) {
    node.Protocol = link.RemoteAddr().Network()
    node.Link = link
    return node
}

func (s *Service) StaticInformation() (info StaticInformation) {
    info.Topology = s.router.TopologyName()
    linkedNode := s.router.GetAllLinkedNode()

    nodes := []string {s.Self().ServeAddress}
    for _, n := range linkedNode {
        nodes = append(nodes, n.ServeAddress)
    }

    info.ConnectedNode = nodes

    return
}

func (s *Service)Self() (node Node) {
    return s.router.Self()
}

func (s *Service) ConnectTo(node Node) (err error) {
    _, err = s.router.ConnectTo(node)
    return
}

func (s *Service) Create(cfg Config) (err error) {
    //create router
    if cfg.Router != nil {
        s.router = cfg.Router
    } else {
        s.router = &Router{}
    }

    //start router
    err = s.router.Start(cfg)
    if err != nil {
        log.Log.Println("create p2p network failed. ",  err)
        return
    }
    s.started = true
    return
}

func (s *Service) Join(seeds []Node, serviceCfg *Config) (err error) {
    if !s.started && serviceCfg != nil{
        err = s.Create(*serviceCfg)
        if err != nil {
            return
        }
    }

    for _, node := range seeds {
        err = s.router.JoinTopology(node)
        if err == nil {
            break
        }

        log.Log.Println("connect to seeds ", node, " failed.")
    }

    if err != nil {
        log.Log.Println("connect to p2p network failed -> seeds : ", seeds)
        return
    }

    s.started = true

    return
}

func (s *Service) Leave() {
    s.router.LeaveTopology()
}

func (s *Service) GetAllLinkedNode() (nodes []Node) {
    return s.router.GetAllLinkedNode()
}

func (s *Service) SendTo(node Node, msg message.Message) (sent int, err error) {
    networkMsg := Message{}
    networkMsg.Message = msg
    return s.router.SendTo(node, networkMsg)
}

func (s *Service) Broadcast(msg message.Message) (err error) {
    return s.router.Broadcast(msg)
}

func (s *Service) RegisterMessageProcessor(msgFamily string, processor MessageProcessor) {
    s.router.RegisterMessageProcessor(msgFamily, processor)
}



