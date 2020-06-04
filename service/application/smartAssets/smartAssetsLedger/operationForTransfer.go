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
	"SealABC/metadata/block"
	"math/big"
)

var bigZero = big.NewInt(0)

func (l Ledger) getBalance(addr []byte, cache txResultCache) (*big.Int, error) {
	addrStr := string(addr)
	if cache[addrStr] != nil {
		return cache[addrStr].val, nil
	}

	balance, err := l.BalanceOf(addr)
	if err == nil {
		cache[addrStr] = &txResultCacheData{
			val:balance,
		}
	}

	return balance, err
}

func (l Ledger) preTransfer(tx Transaction, cache txResultCache, _ block.Entity) ([]StateData, txResultCache, error) {
	if tx.Type != TxType.Transfer.String() {
		return nil, cache, Errors.InvalidTransactionType
	}

	_, err := tx.DataSeal.Verify(tx.getData(), l.CryptoTools.HashCalculator)
	if err != nil {
		return nil, cache, Errors.InvalidParameter.NewErrorWithNewMessage(err.Error())
	}

	fromBalance, err := l.getBalance(tx.From, cache)
	if err != nil {
		return nil, cache, Errors.DBError.NewErrorWithNewMessage(err.Error())
	}

	toBalance, err := l.getBalance(tx.To, cache)
	if err != nil {
		return nil, cache, Errors.DBError.NewErrorWithNewMessage(err.Error())
	}

	if fromBalance.Cmp(bigZero) <= 0 {
		return nil, cache, Errors.InsufficientBalance
	}

	amount, valid := big.NewInt(0).SetString(string(tx.Value), 10)
	if !valid {
		return nil, cache, Errors.InvalidTransferValue
	}

	if amount.Sign() < 0 {
		return nil, cache, Errors.NegativeTransferValue
	} else if amount.Sign() == 0 {
		return nil, cache, Errors.Success
	}

	if fromBalance.Cmp(amount) < 0 {
		return nil, cache, Errors.InsufficientBalance
	}

	fromBalance.Sub(fromBalance, amount)
	toBalance.Add(toBalance, amount)

	cache[string(tx.From)].val = fromBalance
	cache[string(tx.To)].val = toBalance

	statusToChange := []StateData{
		{
			Key: StoragePrefixes.Balance.BuildKey(tx.From),
			Val: fromBalance.Bytes(),
		},

		{
			Key: StoragePrefixes.Balance.BuildKey(tx.To),
			Val: toBalance.Bytes(),
		},
	}

	return statusToChange, cache, Errors.Success
}
