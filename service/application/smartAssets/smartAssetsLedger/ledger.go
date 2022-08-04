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

package smartAssetsLedger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/dataStructure/enum"
	"github.com/SealSC/SealABC/dataStructure/merkleTree"
	"github.com/SealSC/SealABC/dataStructure/state"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/metadata/seal"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
	"math/big"
	"sync"
)

type Ledger struct {
	txPool        *TxPool
	txPoolLimit   int
	clientTxCount map[string]int
	clientTxLimit int

	operateLock sync.RWMutex
	poolLock    sync.Mutex

	genesisAssets BaseAssets

	preActuators   map[string]txPreActuator
	queryActuators map[string]queryActuator

	chain       chainStructure.IChainInterface
	CryptoTools crypto.Tools
	Storage     kvDatabase.IDriver

	storageForEVM contractStorage

	stateDB *state.StateDB

	applicationName string

	genesisRoot []byte
}

func Load() {
	enum.SimpleBuild(&StoragePrefixes)
	enum.SimpleBuild(&StateType)
	enum.SimpleBuild(&TxType)
	enum.SimpleBuild(&QueryTypes)
	enum.SimpleBuild(&QueryParameterFields)
	enum.BuildErrorEnum(&Errors, 1000)
}

func (l *Ledger) SetChain(chain chainStructure.IChainInterface) {
	l.chain = chain

	parentRoot, _ := l.getParentRoot()
	stateDB, err := l.NewState(parentRoot)
	if err != nil {
		panic(err)
		return
	}

	l.stateDB = stateDB

	l.txPool.setChain(chain, l.stateDB)
}

func (l *Ledger) genesisState() {
	batch := l.Storage.NewBatch()

	r, _ := l.stateDB.Commit(batch, true)

	l.genesisRoot = r.Bytes()
	err := l.Storage.BatchWrite(batch)
	if err != nil {
		log.Log.Error("genesis root error")
		return
	}

	return
}

func (l *Ledger) getParentRoot() (root []byte, err error) {
	blockEntity := l.chain.GetLastBlock()
	if blockEntity == nil {
		root = l.genesisRoot
		return
	}
	root, _ = blockEntity.Header.StateRoot[l.applicationName]
	if len(root) == 0 {
		root = l.genesisRoot
	}

	return
}

func (l *Ledger) NewStateAndLoadGenesisAssets(owner []byte, assets BaseAssetsData) (err error) {
	_, exists, err := l.getSystemAssets()
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	if len(owner) == 0 {
		return errors.New("no owner for system assets")
	}

	stateDB, err := l.NewState(common.Hash{}.Bytes())
	if err != nil {
		return err
	}
	l.stateDB = stateDB

	addr := common.BytesToAddress(owner)
	exist := stateDB.Exist(addr)
	if exist {
		return
	}

	supply := assets.Supply
	assets.Supply = "0"

	if len(owner) == 0 {
		return errors.New("no owner for system assets")
	}

	signer, err := l.CryptoTools.SignerGenerator.NewSigner(nil)
	if err != nil {
		return err
	}

	pk := signer.PrivateKeyBytes()
	metaBytes, _ := structSerializer.ToMFBytes(assets)
	metaSeal := seal.Entity{}
	err = metaSeal.Sign(metaBytes, l.CryptoTools, pk)
	if err != nil {
		return err
	}

	assets.Supply = supply
	issuedBytes, _ := structSerializer.ToMFBytes(assets)
	issuedSeal := seal.Entity{}
	err = issuedSeal.Sign(issuedBytes, l.CryptoTools, pk)
	if err != nil {
		return err
	}

	sysAssets := BaseAssets{
		BaseAssetsData: assets,
		IssuedSeal:     issuedSeal,
		MetaSeal:       metaSeal,
	}

	err = l.storeSystemAssets(sysAssets)
	if err != nil {
		return err
	}

	balance, valid := big.NewInt(0).SetString(assets.Supply, 10)
	if !valid {
		return errors.New("invalid assets supply")
	}

	if balance.Sign() <= 0 {
		return errors.New("supply is zero or negative")
	}

	stateDB.AddBalance(addr, balance)
	stateDB.SetNonce(addr, 0)

	l.genesisState()
	return
}

func (l *Ledger) AddTx(req blockchainRequest.Entity) error {
	tx := Transaction{}
	err := json.Unmarshal(req.Data, &tx)
	if err != nil {
		return err
	}

	if tx.Type != req.RequestAction {
		return errors.New("transaction type is not equal to block request action")
	}

	if !bytes.Equal(tx.From, tx.DataSeal.SignerPublicKey) {
		return errors.New("invalid sender")
	}

	valid, err := tx.verify(l.CryptoTools.HashCalculator)
	if !valid {
		return err
	}

	//todo: check balance to ensure client has enough base-assets to pay transaction basic fee

	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	client := string(tx.DataSeal.SignerPublicKey)
	clientTxCount := l.clientTxCount[client]
	if clientTxCount >= l.clientTxLimit {
		return errors.New("reach transaction count limit")
	}

	if l.txPool.len() >= l.txPoolLimit {
		return errors.New("reach transaction pool limit")
	}

	_, exists, _ := l.getTxFromStorage(tx.DataSeal.Hash)
	if exists {
		return errors.New("duplicate history transaction")
	}

	if l.txPool.Get(tx.getCommonHash()) != nil {
		return errors.New("duplicate pending transaction")
	}

	err = l.txPool.addTx(&tx)
	if err != nil {
		return err
	}

	//l.txPoolRecord = append(l.txPoolRecord, tx.getCommonHash())
	l.clientTxCount[client] = clientTxCount + 1

	return nil
}

func (l Ledger) txResultCheck(orgResult TransactionResult, execResult TransactionResult, txHash []byte) error {
	if orgResult.Success != execResult.Success {
		return errors.New(fmt.Sprintf("transaction %x verify failed", txHash))
	}

	if orgResult.ErrorCode != execResult.ErrorCode {
		return errors.New(fmt.Sprintf("transaction %x has different error code", txHash))
	}

	if len(orgResult.NewState) != len(execResult.NewState) {
		return errors.New(fmt.Sprintf("transaction %x has different count state to change", txHash))
	}

	for i, s := range orgResult.NewState {
		if !bytes.Equal(s.Key, execResult.NewState[i].Key) ||
			!bytes.Equal(s.OrgVal, execResult.NewState[i].OrgVal) ||
			!bytes.Equal(s.NewVal, execResult.NewState[i].NewVal) {
			return errors.New(fmt.Sprintf("transaction %x has different state to change", txHash))
		}
	}

	return nil
}

func (l Ledger) PreExecute(txList TransactionList, blk block.Entity) (root []byte, err error) {
	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	txHash := map[string]bool{}
	resultCache := txResultCache{
		CachedBlockGasKey: &txResultCacheData{
			gasLeft: constTransactionGasLimit().Uint64(),
		},

		CachedContractReturnData: &txResultCacheData{
			Data: nil,
		},

		CachedContractCreationAddress: &txResultCacheData{
			Data: nil,
		},
	}

	parentRoot, err := l.getParentRoot()
	if err != nil {
		return
	}
	stateDB, err := l.NewState(parentRoot)
	if err != nil {
		return
	}

	for _, tx := range txList.Transactions {
		hash := string(tx.DataSeal.Hash)
		if _, exists := txHash[hash]; !exists {
			txHash[hash] = true
		} else {
			err = errors.New("duplicate transaction")
			break
		}

		_, err = tx.verify(l.CryptoTools.HashCalculator)
		if err != nil {
			break
		}

		if preExec, exists := l.preActuators[tx.Type]; exists {
			newState, _, execErr := preExec(tx, resultCache, blk)
			txForCheck := Transaction{}
			l.setTxNewState(execErr, newState, &txForCheck)

			checkErr := l.txResultCheck(tx.TransactionResult, txForCheck.TransactionResult, tx.getHash())
			if checkErr != nil {
				err = checkErr
				log.Log.Printf("tx err: %v", err)
				break
			}

			l.setTransactionState(tx, stateDB, nil)
		}
	}

	root = stateDB.IntermediateRoot().Bytes()
	stateRoot := blk.Header.StateRoot[l.applicationName]

	l.removeTransactionsFromPool(txList.Transactions)

	if !bytes.Equal(root, stateRoot) {
		err = fmt.Errorf("invalid merkle root (remote: %x local: %x)", stateRoot, root)
	}

	return
}

func (l *Ledger) removeTransactionsFromPool(txList []Transaction) {
	l.txPool.removeUnenforceable()

	for _, tx := range txList {
		l.txPool.removeTx(tx.getCommonHash())
	}

	for _, tx := range txList {
		client := string(tx.DataSeal.SignerPublicKey)
		if l.clientTxCount[client] > 0 {
			l.clientTxCount[client] -= 1
		}
	}
}

func (l *Ledger) Execute(txList TransactionList, blk block.Entity) (result []byte, root common.Hash, err error) {
	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	batch := l.Storage.NewBatch()

	for _, tx := range txList.Transactions {
		txData, _ := structSerializer.ToMFBytes(tx)
		batch.Put(BuildKey(StoragePrefixes.Transaction, tx.DataSeal.Hash), txData)

		l.setTransactionState(tx, l.stateDB, batch)
	}

	root, err = l.stateDB.Commit(batch, true)
	err = l.Storage.BatchWrite(batch)
	if err != nil {
		return
	}

	return
}

func (l *Ledger) setTransactionState(tx Transaction, state *state.StateDB, batch kvDatabase.Batch) {
	from := common.BytesToAddress(tx.From)

	switch tx.Type {
	case TxType.Transfer.String():
		l.setTransferState(tx.TransactionResult.NewState, state)
	case TxType.CreateContract.String():
		fallthrough
	case TxType.ContractCall.String():
		l.setContractState(tx.TransactionResult.NewState, state, batch)
	}

	state.SetNonce(from, state.GetNonce(from)+1)
}

func (l *Ledger) setContractState(states []StateData, state *state.StateDB, batch kvDatabase.Batch) {
	for _, s := range states {
		switch s.Type {
		case StateType.Balance.String():
			l.setBalance(s.Key, s.NewVal, state)

		case StateType.ContractData.String():

			addr := common.BytesToAddress(s.Key)
			key := common.BytesToHash(s.SubKey)
			value := common.BytesToHash(s.NewVal)

			state.SetState(addr, key, value)
		case StateType.ContractCode.String():

			addr := common.BytesToAddress(s.Key)

			state.SetCode(addr, s.NewVal)
		case StateType.ContractHash.String():
			continue
		default:
			if batch == nil {
				continue
			}
			batch.Put(s.Key, s.NewVal)
		}
	}
}

func (l *Ledger) setTransferState(s []StateData, state *state.StateDB) {
	for _, s := range s {
		l.setBalance(s.Key, s.NewVal, state)
	}
}

func (l *Ledger) setBalance(addr, value []byte, state *state.StateDB) {
	balance := big.NewInt(0)
	balance.SetBytes(value)
	address := common.BytesToAddress(addr)

	state.SetBalance(address, balance)
}

func (l *Ledger) setTxNewState(err error, newState []StateData, tx *Transaction) {
	errEl := err.(enum.ErrorElement)

	if errEl != Errors.Success {
		tx.TransactionResult.Success = false
		tx.TransactionResult.ErrorCode = errEl.Code()
	} else {
		tx.TransactionResult.Success = true
		tx.TransactionResult.NewState = newState
	}
}

func (l *Ledger) GetTransactionsFromPool(blk block.Entity) (txList TransactionList, count uint32, txRoot []byte, stateRoot []byte) {
	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	txs := l.txPool.Pending()
	count = uint32(len(txs))

	if count == 0 {
		return
	}

	resultCache := txResultCache{
		CachedBlockGasKey: &txResultCacheData{
			gasLeft: constTransactionGasLimit().Uint64(),
		},

		CachedContractReturnData: &txResultCacheData{
			Data: nil,
		},

		CachedContractCreationAddress: &txResultCacheData{
			Data: nil,
		},
	}

	parentRoot, err := l.getParentRoot()
	if err != nil {
		return
	}
	stateDB, err := l.NewState(parentRoot)
	if err != nil {
		return
	}

	mt := merkleTree.Tree{}

	for idx, tx := range txs {

		mt.AddHash(tx.DataSeal.Hash)

		if preExec, exists := l.preActuators[tx.Type]; exists {
			newState, _, err := preExec(*tx, resultCache, blk)
			l.setTxNewState(err, newState, tx)
			tx.SequenceNumber = uint32(idx)

			tx.TransactionResult.ReturnData = resultCache[CachedContractReturnData].Data
			tx.TransactionResult.NewAddress = resultCache[CachedContractCreationAddress].address

			l.setTransactionState(*tx, stateDB, nil)
		}

		txList.Transactions = append(txList.Transactions, *tx)
	}

	txRoot, _ = mt.Calculate()
	stateRoot = stateDB.IntermediateRoot().Bytes()

	l.removeTransactionsFromPool(txList.Transactions)
	return
}

func (l *Ledger) DoQuery(req QueryRequest) (interface{}, error) {
	if actuator, exists := l.queryActuators[req.QueryType]; exists {
		return actuator(req)
	}

	return nil, Errors.InvalidQuery
}

func NewLedger(tools crypto.Tools, driver kvDatabase.IDriver, applicationName string) *Ledger {
	l := &Ledger{
		txPool:          NewTxPool(),
		txPoolLimit:     1000,
		clientTxCount:   map[string]int{},
		clientTxLimit:   4,
		operateLock:     sync.RWMutex{},
		poolLock:        sync.Mutex{},
		genesisAssets:   BaseAssets{},
		CryptoTools:     tools,
		Storage:         driver,
		applicationName: applicationName,
	}

	l.storageForEVM.basedLedger = l
	l.preActuators = map[string]txPreActuator{
		TxType.Transfer.String():       l.preTransfer,
		TxType.CreateContract.String(): l.preContractCreation,
		TxType.ContractCall.String():   l.preContractCall,
	}

	l.queryActuators = map[string]queryActuator{
		QueryTypes.BaseAssets.String():   l.queryBaseAssets,
		QueryTypes.Balance.String():      l.queryBalance,
		QueryTypes.Nonce.String():        l.queryNonce,
		QueryTypes.Transaction.String():  l.queryTransaction,
		QueryTypes.OffChainCall.String(): l.contractOffChainCall,
	}

	return l
}
