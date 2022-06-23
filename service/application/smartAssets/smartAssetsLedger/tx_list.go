package smartAssetsLedger

import (
	//"container/heap"
	"container/heap"
	"sort"
)

type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Transactions []*Transaction

func (s Transactions) Len() int { return len(s) }

func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce < s[j].Nonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type txSortedMap struct {
	items map[uint64]*Transaction
	index *nonceHeap
	cache Transactions
}

func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]*Transaction),
		index: new(nonceHeap),
	}
}

func (m *txSortedMap) Get(nonce uint64) *Transaction {
	return m.items[nonce]
}

func (m *txSortedMap) Put(tx *Transaction) {
	nonce := tx.Nonce
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce], m.cache = tx, nil
}

func (m *txSortedMap) Forward(threshold uint64) Transactions {
	var removed Transactions

	for m.index.Len() > 0 && (*m.index)[0] < threshold {
		nonce := heap.Pop(m.index).(uint64)
		removed = append(removed, m.items[nonce])
		delete(m.items, nonce)
	}
	if m.cache != nil {
		m.cache = m.cache[len(removed):]
	}
	return removed
}

func (m *txSortedMap) Filter(filter func(*Transaction) bool) Transactions {
	var removed Transactions

	for nonce, tx := range m.items {
		if filter(tx) {
			removed = append(removed, tx)
			delete(m.items, nonce)
		}
	}
	if len(removed) > 0 {
		*m.index = make([]uint64, 0, len(m.items))
		for nonce := range m.items {
			*m.index = append(*m.index, nonce)
		}
		heap.Init(m.index)

		m.cache = nil
	}
	return removed
}

func (m *txSortedMap) Cap(threshold int) Transactions {
	if len(m.items) <= threshold {
		return nil
	}

	var drops Transactions

	sort.Sort(*m.index)
	for size := len(m.items); size > threshold; size-- {
		drops = append(drops, m.items[(*m.index)[size-1]])
		delete(m.items, (*m.index)[size-1])
	}
	*m.index = (*m.index)[:threshold]
	heap.Init(m.index)

	if m.cache != nil {
		m.cache = m.cache[:len(m.cache)-len(drops)]
	}
	return drops
}

func (m *txSortedMap) Remove(nonce uint64) bool {
	_, ok := m.items[nonce]
	if !ok {
		return false
	}
	for i := 0; i < m.index.Len(); i++ {
		if (*m.index)[i] == nonce {
			heap.Remove(m.index, i)
			break
		}
	}
	delete(m.items, nonce)
	m.cache = nil

	return true
}

func (m *txSortedMap) Ready(start uint64) Transactions {

	if m.index.Len() == 0 || (*m.index)[0] > start {
		return nil
	}

	var ready Transactions
	for next := (*m.index)[0]; m.index.Len() > 0 && (*m.index)[0] == next; next++ {
		ready = append(ready, m.items[next])
		delete(m.items, next)
		heap.Pop(m.index)
	}
	m.cache = nil

	return ready
}

func (m *txSortedMap) Len() int {
	return len(m.items)
}

func (m *txSortedMap) Flatten() Transactions {
	if m.cache == nil {
		m.cache = make(Transactions, 0, len(m.items))
		for _, tx := range m.items {
			m.cache = append(m.cache, tx)
		}
		sort.Sort(TxByNonce(m.cache))
	}

	txs := make(Transactions, len(m.cache))
	copy(txs, m.cache)
	return txs
}

type txList struct {
	strict bool
	txs    *txSortedMap
}

func newTxList(strict bool) *txList {
	return &txList{
		strict: strict,
		txs:    newTxSortedMap(),
	}
}

func (l *txList) Overlaps(tx *Transaction) bool {
	return l.txs.Get(tx.Nonce) != nil
}

func (l *txList) Add(tx *Transaction) (bool, *Transaction) {
	old := l.txs.Get(tx.Nonce)
	l.txs.Put(tx)
	return true, old
}

func (l *txList) Forward(threshold uint64) Transactions {
	return l.txs.Forward(threshold)
}

func (l *txList) Cap(threshold int) Transactions {
	return l.txs.Cap(threshold)
}

func (l *txList) Remove(tx *Transaction) (bool, Transactions) {
	nonce := tx.Nonce
	if removed := l.txs.Remove(nonce); !removed {
		return false, nil
	}
	if l.strict {
		return true, l.txs.Filter(func(tx *Transaction) bool { return tx.Nonce > nonce })
	}
	return true, nil
}

func (l *txList) Ready(start uint64) Transactions {
	return l.txs.Ready(start)
}

func (l *txList) Len() int {
	return l.txs.Len()
}

func (l *txList) Empty() bool {
	return l.Len() == 0
}

func (l *txList) Flatten() Transactions {
	return l.txs.Flatten()
}
