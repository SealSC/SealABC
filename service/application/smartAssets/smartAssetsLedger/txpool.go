package smartAssetsLedger

import (
	"errors"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/dataStructure/state"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"math/big"
	"sync"
)

var (
	ErrNonceTooLow   = errors.New("nonce too low")
	ErrNegativeValue = errors.New("negative value")
)

type TxPool struct {
	chain chainStructure.IChainInterface

	pending map[common.Address]*txList
	//queue   map[common.Address]*txList
	all map[common.Hash]*Transaction

	currentState *state.StateDB
	pendingState *state.ManagedState

	mu sync.RWMutex
}

func NewTxPool() *TxPool {
	pool := &TxPool{
		mu: sync.RWMutex{},

		pending: make(map[common.Address]*txList),
		//queue:   make(map[common.Address]*txList),
		all: make(map[common.Hash]*Transaction),
	}

	return pool
}

func (pool *TxPool) setChain(chain chainStructure.IChainInterface, stateDB *state.StateDB) {
	pool.chain = chain
	pool.currentState = stateDB
	pool.pendingState = state.ManageState(stateDB)
}

func (pool *TxPool) len() int {
	return len(pool.all)
}

func (pool *TxPool) Get(hash common.Hash) *Transaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.all[hash]
}

func (pool *TxPool) addTx(tx *Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	err := pool.add(tx)
	if err != nil {
		return err
	}

	return nil
}

func (pool *TxPool) add(tx *Transaction) error {
	hash := tx.getCommonHash()

	err := pool.validateTx(tx)
	if err != nil {
		return err
	}

	pool.promoteTx(common.BytesToAddress(tx.From), hash, tx)

	return nil
}

func (pool *TxPool) validateTx(tx *Transaction) error {
	amount, _ := big.NewInt(0).SetString(string(tx.Value), 10)
	if amount.Sign() < 0 {
		return ErrNegativeValue
	}

	from := common.BytesToAddress(tx.From)
	if pool.currentState.GetNonce(from) > tx.Nonce {
		return ErrNonceTooLow
	}

	return nil
}

func (pool *TxPool) promoteTx(addr common.Address, hash common.Hash, tx *Transaction) {
	if pool.pending[addr] == nil {
		pool.pending[addr] = newTxList(true)
	}
	list := pool.pending[addr]

	_, old := list.Add(tx)

	if old != nil {
		delete(pool.all, old.getCommonHash())
	}

	if pool.all[hash] == nil {
		pool.all[hash] = tx
	}

	pool.pendingState.SetNonce(addr, tx.Nonce+1)
}

func (pool *TxPool) removeTx(hash common.Hash) {
	tx, ok := pool.all[hash]
	if !ok {
		return
	}
	addr := common.BytesToAddress(tx.From)

	delete(pool.all, hash)

	if pending := pool.pending[addr]; pending != nil {
		if removed, _ := pending.Remove(tx); removed {
			if pending.Empty() {
				delete(pool.pending, addr)
			} else {
				//for _, tx := range invalids {
				//
				//}
			}

			if nonce := tx.Nonce; pool.pendingState.GetNonce(addr) > nonce {
				pool.pendingState.SetNonce(addr, nonce)
			}
			return
		}
	}

	//if future := pool.queue[addr]; future != nil {
	//	future.Remove(tx)
	//	if future.Empty() {
	//		delete(pool.queue, addr)
	//	}
	//}
}

func (pool *TxPool) removeUnenforceable() {
	for addr, list := range pool.pending {
		nonce := pool.currentState.GetNonce(addr)

		for _, tx := range list.Forward(nonce) {
			hash := tx.getCommonHash()
			delete(pool.all, hash)
		}

		if list.Empty() {
			delete(pool.pending, addr)
		}
	}
}

func (pool *TxPool) Pending() Transactions {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	var pending Transactions
	for _, list := range pool.pending {
		pending = append(pending, list.Flatten()...)
	}
	return pending
}
