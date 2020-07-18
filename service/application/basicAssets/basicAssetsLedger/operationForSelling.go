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

package basicAssetsLedger

import (
	"bytes"
	"encoding/json"
	"errors"
)

const MarketAddress = "MarketAddress"

type SellingOperationResult struct {
	UnspentListWithBalance
	SellingData
}

func (l *Ledger) sellingDataVerify(tx Transaction) (data SellingData, err error)  {
	err = json.Unmarshal(tx.ExtraData, &data)
	if err != nil {
		return
	}

	if bytes.Equal(data.PaymentAssets, data.SellingAssets) {
		return data, errors.New("payment assets must not as same as selling assets")
	}

	_, err = l.localAssetsFromHash(data.PaymentAssets)
	if err != nil {
		return
	}

	return
}

func (l *Ledger) verifyStartSelling(tx Transaction) (ret interface{}, err error) {
	if len(tx.Input) != 1 || len(tx.Output) != 0 {
		return nil, errors.New("invalid input or output count")
	}

	sellingData, err := l.sellingDataVerify(tx)
	if  err != nil {
		return
	}

	if sellingData.Amount < 1 {
		return nil, errors.New("invalid amount")
	}

	if !bytes.Equal(sellingData.SellingAssets, tx.Assets.MetaSeal.Hash) {
		return nil, errors.New("invalid assets")
	}

	if !bytes.Equal(sellingData.Seller, tx.Seal.SignerPublicKey) {
		return nil, errors.New("invalid owner")
	}

	ref := tx.Input[0]
	key := l.buildUnspentStorageKey(tx.Seal.SignerPublicKey, tx.Assets.getUniqueHash(), ref.Transaction, ref.OutputIndex)
	unspent, err := l.getUnspent(key)
	if err != nil {
		return
	}

	return []Unspent{unspent}, nil
}

func (l *Ledger) verifyStopSelling(tx Transaction) (ret interface{}, err error) {
	if len(tx.Input) != 0 || len(tx.Output) != 0 {
		return nil, errors.New("invalid input or output count")
	}
	sellingData, err := l.sellingDataVerify(tx)
	if  err != nil {
		return
	}

	key := l.buildUnspentStorageKey([]byte(MarketAddress), sellingData.SellingAssets, sellingData.Transaction, 0)
	unspent, dbErr := l.getUnspent(key)
	if dbErr != nil {
		err = errors.New("get Unspent failed: " + dbErr.Error())
		return
	}

	if !bytes.Equal(unspent.Singer, tx.Seal.SignerPublicKey) {
		return nil, errors.New("not assets owner")
	}

	return []Unspent{unspent}, nil
}

func (l *Ledger) verifyBuyAssets(tx Transaction) (ret interface{}, err error) {
	inCount := len(tx.Input)
	if inCount < 2 {
		return nil, errors.New("invalid input or output count")
	}

	target := tx.Input[inCount - 1]
	sellingKey := l.buildAssetsSellingKey(target.Transaction)
	data, err := l.Storage.Get(sellingKey)
	if err != nil {
		return nil, err
	}

	if !data.Exists {
		return nil, errors.New("not selling")
	}

	sellingData := SellingData{}
	_ = json.Unmarshal(data.Data, &sellingData)

	inAmount := uint64(0)
	var uList []Unspent

	for i:=0; i < inCount - 1; i++ {
		key := l.buildUnspentStorageKey(tx.Seal.SignerPublicKey, tx.Assets.getUniqueHash(), tx.Input[i].Transaction, tx.Input[i].OutputIndex)
		unspent, dbErr := l.getUnspent(key)
		if dbErr != nil {
			err = errors.New("get Unspent failed: " + dbErr.Error())
			break
		}

		inAmount += unspent.Value
		uList = append(uList, unspent)
	}

	if err != nil {
		return
	}

	outAmount := uint64(0)
	for _, out := range tx.Output {
		outAmount += out.Value
	}

	if inAmount != outAmount {
		return nil, errors.New("input amount not equal output")
	}

	return uList, nil
}

func (l *Ledger) confirmStartSelling(tx Transaction) (ret interface{}, err error) {
	l.operateLock.Lock()
	defer l.operateLock.Unlock()

	//tx in this phase was verified, get unspent list directly
	usList, inAmount, err := l.getUnspentListFromTransaction(tx)
	if err != nil {
		return
	}

	localAssets, _ := l.localAssetsFromHash(tx.Assets.getUniqueHash())

	tx.Output = []UTXOOutput{
		{
			To:    []byte(MarketAddress),
			Value: inAmount,
		},
	}

	ul, err := l.saveUnspent(localAssets, tx, usList)

	if err == nil {
		l.updateDoubleSpentCache(usList)
	}

	sellingData := SellingData{}
	_ = json.Unmarshal(tx.ExtraData, &sellingData)

	sellingData.Transaction = tx.Seal.Hash
	sellingJson, _ := json.Marshal(sellingData)
	err = l.storeSellingData(tx.Seal.Hash, sellingJson)
	if err != nil {
		return
	}



	ret = SellingOperationResult {
		UnspentListWithBalance: ul,
		SellingData:            sellingData,
	}

	return
}

func (l *Ledger) confirmStopSelling(tx Transaction) (ret interface{}, err error) {
	l.operateLock.Lock()
	defer l.operateLock.Unlock()

	sellingData := SellingData{}
	_ = json.Unmarshal(tx.ExtraData, &sellingData)

	key := l.buildUnspentStorageKey([]byte(MarketAddress), sellingData.SellingAssets, sellingData.Transaction, 0)
	unspent, err := l.getUnspent(key)
	if err != nil {
		err = errors.New("get Unspent failed: " + err.Error())
		return
	}

	localAssets, _ := l.localAssetsFromHash(sellingData.SellingAssets)

	tx.Output = []UTXOOutput{
		{
			To:    tx.Seal.SignerPublicKey,
			Value: unspent.Value,
		},
	}

	uList := []Unspent{unspent}

	ul, err := l.saveUnspent(localAssets, tx, uList)

	if err == nil {
		l.updateDoubleSpentCache(uList)
	}

	_ = l.deleteSellingData(sellingData.Transaction)

	ret = SellingOperationResult {
		UnspentListWithBalance: ul,
		SellingData:            sellingData,
	}
	return
}

func (l *Ledger) confirmBuyAssets(tx Transaction) (ret interface{}, err error) {
	l.operateLock.Lock()
	defer l.operateLock.Unlock()

	inCount := len(tx.Input)
	target := tx.Input[inCount - 1]

	sellingKey := l.buildAssetsSellingKey(target.Transaction)
	data, err := l.Storage.Get(sellingKey)
	if err != nil {
		return nil, err
	}

	sellingData := SellingData{}
	_ = json.Unmarshal(data.Data, &sellingData)

	var uList []Unspent

	for i:=0; i < inCount - 1; i++ {
		key := l.buildUnspentStorageKey(tx.Seal.SignerPublicKey, tx.Assets.getUniqueHash(), tx.Input[i].Transaction, tx.Input[i].OutputIndex)
		unspent, dbErr := l.getUnspent(key)
		if dbErr != nil {
			err = errors.New("get Unspent failed: " + dbErr.Error())
			return
		}

		uList = append(uList, unspent)
	}

	key := l.buildUnspentStorageKey([]byte(MarketAddress), sellingData.SellingAssets, target.Transaction, target.OutputIndex)
	unspent, err := l.getUnspent(key)
	if err != nil {
		err = errors.New("get Unspent failed: " + err.Error())
		return
	}

	paymentAssets, _ := l.localAssetsFromHash(sellingData.PaymentAssets)

	payUl, err := l.saveUnspent(paymentAssets, tx, uList)
	if err == nil {
		l.updateDoubleSpentCache(uList)
	}

	sellAssets, _ := l.localAssetsFromHash(sellingData.SellingAssets)
	tx.Input = []UTXOInput{}
	tx.Output =[]UTXOOutput{{
		To:    tx.Seal.SignerPublicKey,
		Value: sellingData.Amount,
	}}

	buyUl, err := l.saveUnspent(sellAssets, tx, []Unspent{unspent})

	if err == nil {
		l.updateDoubleSpentCache(uList)
	}

	payUl.BalanceList = append(payUl.BalanceList, buyUl.BalanceList...)
	payUl.UnspentList = append(payUl.UnspentList, buyUl.UnspentList...)

	_ = l.deleteSellingData(target.Transaction)
	if err != nil {
		return
	}

	if sellAssets.Type == uint32(AssetsTypes.Copyright.Int()) {
		err = l.copyrightOwnerTransfer(sellAssets.getUniqueHash(), tx.Seal.SignerPublicKey)
	}

	ret = SellingOperationResult{
		UnspentListWithBalance: payUl,
		SellingData:            sellingData,
	}

	return
}
