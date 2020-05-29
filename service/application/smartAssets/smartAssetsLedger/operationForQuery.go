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
	"encoding/hex"
	"encoding/json"
)

func (l Ledger) queryBaseAssets(_ QueryRequest) (interface{}, error) {
	return l.genesisAssets, nil
}

func (l Ledger) queryBalance(req QueryRequest) (interface{}, error) {
	hexStr := req.Parameter[QueryParameterFields.Address.String()]
	if hexStr == "" {
		return nil, Errors.InvalidParameter
	}

	addr, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, Errors.InvalidParameter.NewErrorWithNewMessage(err.Error())
	}

	balance, err := l.balanceOf(addr, l.genesisAssets.getHash())
	if err != nil {
		return nil, Errors.DBError.NewErrorWithNewMessage(err.Error())
	}
	
	return balance.String(), nil
}

func (l Ledger) queryTransaction(req QueryRequest) (interface{}, error) {
	param := req.Parameter[QueryParameterFields.TxHash.String()]
	if param == "" {
		return nil, Errors.InvalidParameter
	}

	txHash, err := hex.DecodeString(param)
	if err != nil {
		return nil, Errors.InvalidParameter.NewErrorWithNewMessage(err.Error())
	}

	tx, _, err := l.getTxFromStorage(txHash)

	return tx, err
}

func (l Ledger) contractOffChainCall(req QueryRequest) (interface{}, error) {
	txJson := req.Parameter[QueryParameterFields.Data.String()]
	if txJson == "" {
		return nil, Errors.InvalidParameter
	}

	blk := l.chain.GetLastBlock()
	tx := Transaction{}
	err := json.Unmarshal([]byte(txJson), &tx)
	if err != nil {
		return nil, Errors.InvalidParameter.NewErrorWithNewMessage(err.Error())
	}

	_, cache, err := l.preContractCall(tx, nil, *blk)

	return cache, err
}
