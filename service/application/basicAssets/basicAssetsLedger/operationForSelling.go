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
	"SealABC/log"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
)

const MarketAddress = "MarketAddress"

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
	log.Log.Warn(hex.EncodeToString(key))
	unspent, err := l.getUnspent(key)
	if err != nil {
		return
	}

	return []Unspent{unspent}, nil
}

func (l *Ledger) verifyStopSelling(tx Transaction) (ret interface{}, err error) {
	if len(tx.Input) != 1 || len(tx.Output) != 0 {
		return nil, errors.New("invalid input or output count")
	}

	inTx := tx.Input[0]
	key := l.buildUnspentStorageKey([]byte(MarketAddress), tx.Assets.getUniqueHash(), inTx.Transaction, inTx.OutputIndex)
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

	outAmount += sellingData.Price

	if inAmount != outAmount {
		log.Log.Warn(inAmount, ":", outAmount)
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

	ret, err = l.saveUnspent(localAssets, tx, usList)

	if err == nil {
		l.updateDoubleSpentCache(usList)
	}

	err = l.storeSellingData(tx.Seal.Hash, tx.ExtraData)
	if err != nil {
		return
	}

	return
}

func (l *Ledger) confirmStopSelling(tx Transaction) (ret interface{}, err error) {
	l.operateLock.Lock()
	defer l.operateLock.Unlock()

	inTx := tx.Input[0]
	key := l.buildUnspentStorageKey([]byte(MarketAddress), tx.Assets.getUniqueHash(), inTx.Transaction, inTx.OutputIndex)
	unspent, err := l.getUnspent(key)
	if err != nil {
		err = errors.New("get Unspent failed: " + err.Error())
		return
	}

	localAssets, _ := l.localAssetsFromHash(tx.Assets.getUniqueHash())

	tx.Output = []UTXOOutput{
		{
			To:    tx.Seal.SignerPublicKey,
			Value: unspent.Value,
		},
	}

	uList := []Unspent{unspent}

	ret, err = l.saveUnspent(localAssets, tx, uList)

	if err == nil {
		l.updateDoubleSpentCache(uList)
	}

	_ = l.deleteSellingData(inTx.Transaction)
	if err != nil {
		return
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

	tx.Output = append(tx.Output, UTXOOutput{
		To:    sellingData.Seller,
		Value: sellingData.Price,
	})

	uList = append(uList, unspent)
	ret, err = l.saveUnspent(paymentAssets, tx, uList)

	sellAssets, _ := l.localAssetsFromHash(sellingData.SellingAssets)
	tx.Input = []UTXOInput{}
	tx.Output =[]UTXOOutput{{
		To:    tx.Seal.SignerPublicKey,
		Value: sellingData.Amount,
	}}

	ret, _ = l.saveUnspent(sellAssets, tx, []Unspent{})

	if err == nil {
		l.updateDoubleSpentCache(uList)
	}

	_ = l.deleteSellingData(target.Transaction)
	if err != nil {
		return
	}

	return
}
