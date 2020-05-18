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

//type TopologyRegister func(IProcessor) (error)

type ITopology interface {
    Name() string

    MountTo(router IRouter)
    BuildNodeID(node Node) string

    InterestedMessage(msg Message) (interested bool)
    MessageProcessor(msg Message, link ILink)

    Join(seed LinkNode) (err error)
    Leave()

    SetLocalNode(node LinkNode)
    GetLink(node Node) (link ILink, err error)
    GetAllNodes() []LinkNode

    AddLink(link ILink)
    RemoveLink(link ILink)
}

type directConnect struct{
    LocalNode LinkNode
}

func (t *directConnect) Name() string {
    return "direct connect"
}

func (t *directConnect)MountTo(router IRouter) {
    return
}

func (t *directConnect)BuildNodeID(node Node) string {
    return node.ServeAddress
}

func (t *directConnect)InterestedMessage(msg Message) (interested bool)  {
    return false
}

func (t *directConnect)MessageProcessor(msg Message, link ILink) {
    return
}

func (t *directConnect)Join(node LinkNode) (err error) {
    return
}

func (t *directConnect)Leave() {
    return
}


func (t *directConnect)SetLocalNode(node LinkNode) {
    t.LocalNode = node
}

func (t *directConnect)GetLocalNode() (node LinkNode) {
    return t.LocalNode
}

func (t *directConnect)GetLink(node Node) (link ILink, err error) {
    return
}

func (t *directConnect)GetAllNodes() (all []LinkNode) {
    return
}

func (t *directConnect)AddLink(link ILink) {

}

func (t *directConnect)RemoveLink(link ILink) {

}