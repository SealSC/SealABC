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
	"SealEVM/evmInt256"
)

func (l Ledger) preContractCall(tx Transaction, cache txResultCache, blk block.Entity) ([]StateData, txResultCache, error) {
	if tx.Type != TxType.ContractCall.String() {
		return  nil, cache, Errors.InvalidTransactionType
	}

	if len(tx.To) == 0 {
		return nil, nil, Errors.InvalidContractCreationAddress
	}

	initGas := cache[CachedBlockGasKey].gasLeft
	evm, _, err := l.newEVM(tx, nil, blk, evmInt256.New(int64(initGas)))
	if err != nil {
		return nil, cache, err
	}

	execErr := Errors.Success
	ret, err := evm.ExecuteContract(true)
	if err != nil {
		execErr = Errors.ContractExecuteFailed.NewErrorWithNewMessage(err.Error())
	}
	newState := l.newStateFromEVMResult(ret, cache)

	gasCost := initGas - ret.GasLeft
	cache[CachedBlockGasKey].gasLeft -= gasCost
	cache[CachedContractReturnData].data = ret.ResultData
	return newState, cache, execErr
}
