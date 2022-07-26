package trie

const (
	NodeTypeFull  = "Full"
	NodeTypeShort = "Short"
	NodeTypeHash  = "Hash"
	NodeTypeValue = "Value"
)

type BatchWriter interface {
	Put(key, value []byte)
}
