package trie

type BatchWriter interface {
	Put(key, value []byte)
}
