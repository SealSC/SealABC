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
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealEVM/evmInt256"
	"github.com/SealSC/SealEVM/opcodes"
)

func (l *Ledger) preContractCreation(tx Transaction, cache txResultCache, blk block.Entity) ([]StateData, txResultCache, error) {
	if tx.Type != TxType.CreateContract.String() {
		return nil, cache, Errors.InvalidTransactionType
	}

	if len(tx.To) != 0 {
		return nil, nil, Errors.InvalidContractCreationAddress
	}

	initGas := cache[CachedBlockGasKey].gasLeft
	evm, contract, _ := l.newEVM(tx, nil, blk, evmInt256.New(int64(initGas)))

	ret, err := evm.ExecuteContract(true)
	newState := l.newStateFromEVMResult(ret, cache)

	gasCost := initGas - ret.GasLeft
	cache[CachedBlockGasKey].gasLeft -= gasCost

	if err == nil {
		if ret.ExitOpCode == opcodes.REVERT {
			err = Errors.ContractExecuteRevert
			return nil, nil, err
		}

		contractAddr := contract.Namespace.Bytes()
		newState = append(newState,
			StateData{
				Key:    BuildKey(StoragePrefixes.ContractCode, contractAddr),
				NewVal: ret.ResultData,
			},
			StateData{
				Key:    BuildKey(StoragePrefixes.ContractHash, contractAddr),
				NewVal: contract.Hash.Bytes(),
			},
		)
	} else {
		return nil, nil, Errors.ContractCreationFailed.NewErrorWithNewMessage(err.Error())
	}

	cache[CachedContractCreationAddress].address = common.BytesToAddress(contract.Namespace.Bytes())
	cache[CachedContractReturnData].Data = ret.ResultData
	return newState, cache, Errors.Success
}
