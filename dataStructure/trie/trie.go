package trie

import (
	"bytes"
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	emptyState common.Hash
)

var (
	cacheMissCounter   = metrics.NewRegisteredCounter("trie/cachemiss", nil)
	cacheUnloadCounter = metrics.NewRegisteredCounter("trie/cacheunload", nil)
)

type Trie struct {
	root         node
	db           kvDatabase.IDriver
	originalRoot common.Hash

	cacheGen   uint16
	cacheLimit uint16
}

func New(root common.Hash, db kvDatabase.IDriver) (*Trie, error) {
	trie := &Trie{db: db, originalRoot: root}
	if (root != common.Hash{}) && root != emptyRoot {
		if db == nil {
			panic("trie.New: cannot use existing root without a database")
		}
		rootNode, err := trie.resolveHash(root[:], nil)
		if err != nil {
			return nil, err
		}
		trie.root = rootNode
	}
	return trie, nil
}

func (t *Trie) NodeIterator(start []byte) NodeIterator {
	return newNodeIterator(t, start)
}

func (t *Trie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		log.Log.Errorf("Unhandled trie error: %v", err)
	}
	return res
}

func (t *Trie) TryGet(key []byte) ([]byte, error) {
	key = keyBytesToHex(key)
	value, newRoot, didResolve, err := t.tryGet(t.root, key, 0)
	if err == nil && didResolve {
		t.root = newRoot
	}
	return value, err
}

func (t *Trie) tryGet(origNode node, key []byte, pos int) (value []byte, newnode node, didResolve bool, err error) {
	switch n := (origNode).(type) {
	case nil:
		return nil, nil, false, nil
	case valueNode:
		return n, n, false, nil
	case *shortNode:
		if len(key)-pos < len(n.Key) || !bytes.Equal(n.Key, key[pos:pos+len(n.Key)]) {
			// key not found in trie
			return nil, n, false, nil
		}

		value, newnode, didResolve, err = t.tryGet(n.Val, key, pos+len(n.Key))

		if err == nil && didResolve {
			n = n.copy()
			n.Val = newnode
			n.flags.gen = t.cacheGen
		}

		return value, n, didResolve, err
	case *fullNode:
		value, newnode, didResolve, err = t.tryGet(n.Children[key[pos]], key, pos+1)
		if err == nil && didResolve {
			n = n.copy()
			n.flags.gen = t.cacheGen
			n.Children[key[pos]] = newnode
		}
		return value, n, didResolve, err
	case hashNode:
		child, err := t.resolveHash(n, key[:pos])
		if err != nil {
			return nil, n, true, err
		}
		value, newNode, _, err := t.tryGet(child, key, pos)
		return value, newNode, true, err
	default:
		panic(fmt.Sprintf("%T: invalid node: %v", origNode, origNode))
	}
}

func (t *Trie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		log.Log.Errorf("Unhandled trie error: %v", err)
	}
}

func (t *Trie) TryUpdate(key, value []byte) error {
	k := keyBytesToHex(key)
	if len(value) != 0 {
		_, n, err := t.insert(t.root, nil, k, valueNode(value))
		if err != nil {
			return err
		}
		t.root = n
	} else {
		_, n, err := t.delete(t.root, nil, k)
		if err != nil {
			return err
		}
		t.root = n
	}
	return nil
}

func (t *Trie) insert(n node, prefix, key []byte, value node) (bool, node, error) {
	if len(key) == 0 {
		if v, ok := n.(valueNode); ok {
			return !bytes.Equal(v, value.(valueNode)), value, nil
		}
		return true, value, nil
	}
	switch n := n.(type) {
	case *shortNode:
		matchLen := prefixLen(key, n.Key)
		if matchLen == len(n.Key) {
			dirty, nn, err := t.insert(n.Val, append(prefix, key[:matchLen]...), key[matchLen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &shortNode{n.Key, nn, t.newFlag()}, nil
		}
		branch := &fullNode{flags: t.newFlag()}
		var err error
		_, branch.Children[n.Key[matchLen]], err = t.insert(nil, append(prefix, n.Key[:matchLen+1]...), n.Key[matchLen+1:], n.Val)
		if err != nil {
			return false, nil, err
		}
		_, branch.Children[key[matchLen]], err = t.insert(nil, append(prefix, key[:matchLen+1]...), key[matchLen+1:], value)
		if err != nil {
			return false, nil, err
		}
		if matchLen == 0 {
			return true, branch, nil
		}
		return true, &shortNode{key[:matchLen], branch, t.newFlag()}, nil

	case *fullNode:
		dirty, nn, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.flags = t.newFlag()
		n.Children[key[0]] = nn
		return true, n, nil

	case nil:
		return true, &shortNode{key, value, t.newFlag()}, nil

	case hashNode:
		rn, err := t.resolveHash(n, prefix)
		if err != nil {
			return false, nil, err
		}
		dirty, nn, err := t.insert(rn, prefix, key, value)
		if !dirty || err != nil {
			return false, rn, err
		}
		return true, nn, nil

	default:
		log.Log.Errorf("%T: invalid node: %v", n, n)
		return false, nil, nil
	}
}

func (t *Trie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil {
		log.Log.Errorf("Unhandled trie error: %v", err)
	}
}

func (t *Trie) TryDelete(key []byte) error {
	k := keyBytesToHex(key)
	_, n, err := t.delete(t.root, nil, k)
	if err != nil {
		return err
	}
	t.root = n
	return nil
}

func (t *Trie) delete(n node, prefix, key []byte) (bool, node, error) {
	switch n := n.(type) {
	case *shortNode:
		matchLen := prefixLen(key, n.Key)
		if matchLen < len(n.Key) {
			return false, n, nil
		}
		if matchLen == len(key) {
			return true, nil, nil
		}

		dirty, child, err := t.delete(n.Val, append(prefix, key[:len(n.Key)]...), key[len(n.Key):])
		if !dirty || err != nil {
			return false, n, err
		}
		switch child := child.(type) {
		case *shortNode:
			return true, &shortNode{concat(n.Key, child.Key...), child.Val, t.newFlag()}, nil
		default:
			return true, &shortNode{n.Key, child, t.newFlag()}, nil
		}

	case *fullNode:
		dirty, nn, err := t.delete(n.Children[key[0]], append(prefix, key[0]), key[1:])
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.flags = t.newFlag()
		n.Children[key[0]] = nn

		pos := -1
		for i, cld := range n.Children {
			if cld != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			if pos != 16 {
				cNode, err := t.resolve(n.Children[pos], prefix)
				if err != nil {
					return false, nil, err
				}
				if cNode, ok := cNode.(*shortNode); ok {
					k := append([]byte{byte(pos)}, cNode.Key...)
					return true, &shortNode{k, cNode.Val, t.newFlag()}, nil
				}
			}

			return true, &shortNode{[]byte{byte(pos)}, n.Children[pos], t.newFlag()}, nil
		}

		return true, n, nil

	case valueNode:
		return true, nil, nil

	case nil:
		return false, nil, nil

	case hashNode:
		rn, err := t.resolveHash(n, prefix)
		if err != nil {
			return false, nil, err
		}
		dirty, nn, err := t.delete(rn, prefix, key)
		if !dirty || err != nil {
			return false, rn, err
		}
		return true, nn, nil

	default:
		panic(fmt.Sprintf("%T: invalid node: %v (%v)", n, n, key))
	}
}

func (t *Trie) Root() []byte { return t.Hash().Bytes() }

func (t *Trie) Hash() common.Hash {
	hash, cached, _ := t.hashRoot(nil)
	t.root = cached
	return common.BytesToHash(hash.(hashNode))
}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}

func (t *Trie) resolve(n node, prefix []byte) (node, error) {
	if n, ok := n.(hashNode); ok {
		return t.resolveHash(n, prefix)
	}
	return n, nil
}

func (t *Trie) hashRoot(bw BatchWriter) (node, node, error) {
	if t.root == nil {
		return hashNode(emptyRoot.Bytes()), nil, nil
	}
	h := newHasher(t.cacheGen, t.cacheLimit)
	defer returnHasherToPool(h)
	return h.hash(t.root, bw, true)
}

func (t *Trie) SetCacheLimit(l uint16) {
	t.cacheLimit = l
}

func (t *Trie) newFlag() nodeFlag {
	return nodeFlag{dirty: true, gen: t.cacheGen}
}

func (t *Trie) resolveHash(n hashNode, prefix []byte) (node, error) {
	cacheMissCounter.Inc(1)

	enc, err := t.db.Get(n)
	if err != nil || !enc.Exists {
		return nil, &MissingNodeError{NodeHash: common.BytesToHash(n), Path: prefix}
	}
	dec := mustDecodeNode(n, enc.Data, t.cacheGen)
	return dec, nil
}

func mustDecodeNode(hash, buf []byte, cacheGen uint16) node {
	n, err := decodeNode(hash, buf, cacheGen)
	if err != nil {
		panic(fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

func (t *Trie) CommitTo(bw BatchWriter) (root common.Hash, err error) {
	hash, cached, err := t.hashRoot(bw)
	if err != nil {
		return common.Hash{}, err
	}
	t.root = cached
	t.cacheGen++
	return common.BytesToHash(hash.(hashNode)), nil
}
