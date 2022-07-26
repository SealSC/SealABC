package trie

import "fmt"

func (n *fullNode) String() string  { return n.fString("") }
func (n *shortNode) String() string { return n.fString("") }
func (n hashNode) String() string   { return n.fString("") }
func (n valueNode) String() string  { return n.fString("") }

func (n *fullNode) fString(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fString(ind+"  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}
func (n *shortNode) fString(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fString(ind+"  "))
}
func (n hashNode) fString(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}
func (n valueNode) fString(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}
