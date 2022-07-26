package trie

import (
	"fmt"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"strings"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
	fString(string) string
	cache() (hashNode, bool)
	canUnload(cacheGen, cacheLimit uint16) bool
	encodeToPN() persistenceNode
}

type (
	fullNode struct {
		Children [17]node
		flags    nodeFlag
	}
	shortNode struct {
		Key   []byte
		Val   node
		flags nodeFlag
	}
	hashNode  []byte
	valueNode []byte
)

func (n *fullNode) canUnload(gen, limit uint16) bool  { return n.flags.canUnload(gen, limit) }
func (n *shortNode) canUnload(gen, limit uint16) bool { return n.flags.canUnload(gen, limit) }
func (n hashNode) canUnload(uint16, uint16) bool      { return false }
func (n valueNode) canUnload(uint16, uint16) bool     { return false }

func (n *fullNode) cache() (hashNode, bool)  { return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() (hashNode, bool) { return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() (hashNode, bool)   { return nil, true }
func (n valueNode) cache() (hashNode, bool)  { return nil, true }

func (n *fullNode) copy() *fullNode   { cn := *n; return &cn }
func (n *shortNode) copy() *shortNode { cn := *n; return &cn }

func (n *fullNode) encodeToPN() (pn persistenceNode) {
	pn.NodeType = NodeTypeFull
	nodes := make([]persistenceNode, len(n.Children))
	for i, node := range n.Children {
		if node == nil {
			continue
		}
		nodes[i] = node.encodeToPN()
	}

	pn.Nodes = nodes
	return
}
func (n *shortNode) encodeToPN() (pn persistenceNode) {
	pn.NodeType = NodeTypeShort
	node := n.Val.encodeToPN()
	pn.Key = compactToHex(n.Key)
	pn.Nodes = []persistenceNode{node}
	return
}
func (n hashNode) encodeToPN() (pn persistenceNode) {
	pn.NodeType = NodeTypeHash
	pn.Data = n
	return
}
func (n valueNode) encodeToPN() (pn persistenceNode) {
	pn.NodeType = NodeTypeValue
	pn.Data = n
	return
}

type nodeFlag struct {
	hash  hashNode
	gen   uint16
	dirty bool
}

func (n *nodeFlag) canUnload(cacheGen, cacheLimit uint16) bool {
	return !n.dirty && cacheGen-n.gen >= cacheLimit
}

type persistenceNode struct {
	NodeType string
	Nodes    []persistenceNode
	Key      []byte
	Data     []byte
}

func pnToNode(hash []byte, pn persistenceNode, cacheGen uint16) (n node) {
	switch pn.NodeType {
	case NodeTypeFull:
		child := &fullNode{flags: nodeFlag{hash: hash, gen: cacheGen}}
		nodes := [17]node{}
		for i, node := range pn.Nodes {
			nodes[i] = pnToNode(hash, node, cacheGen)
		}
		child.Children = nodes
		n = child
	case NodeTypeShort:
		node := &shortNode{}
		node.flags = nodeFlag{hash: hash, gen: cacheGen}
		node.Key = pn.Key
		child := pn.Nodes[0]
		node.Val = pnToNode(hash, child, cacheGen)
		n = node
	case NodeTypeHash:
		if len(pn.Data) == 0 {
			return nil
		}
		return hashNode(pn.Data)
	case NodeTypeValue:
		if len(pn.Data) == 0 {
			return nil
		}
		return valueNode(pn.Data)
	}
	return
}

func encodeNode(n node) ([]byte, error) {
	pn := n.encodeToPN()
	return structSerializer.ToMFBytes(pn)
}

func decodeNode(hash, buf []byte, cacheGen uint16) (node, error) {
	pn := persistenceNode{}
	err := structSerializer.FromMFBytes(buf, &pn)
	n := pnToNode(hash, pn, cacheGen)

	return n, err
}

type decodeError struct {
	what  error
	stack []string
}

func wrapError(err error, ctx string) error {
	if err == nil {
		return nil
	}
	if decErr, ok := err.(*decodeError); ok {
		decErr.stack = append(decErr.stack, ctx)
		return decErr
	}
	return &decodeError{err, []string{ctx}}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}
