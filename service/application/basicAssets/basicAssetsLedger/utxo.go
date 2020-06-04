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
    "SealABC/storage/db/dbInterface/kvDatabase"
    "encoding/binary"
    "encoding/json"
    "errors"
)

type UTXOInput struct {
    Transaction  []byte
    OutputIndex  uint64 `json:",string"`
}

type UTXOOutput struct {
    To      []byte
    Value   uint64 `json:",string"`
}

type Unspent struct {
    Owner       []byte
    AssetsHash  []byte
    Transaction []byte
    OutputIndex uint64 `json:",string"`
    Value       uint64 `json:",string"`
}

type Balance struct {
    Address []byte
    Assets  Assets
    Amount  uint64
}

type UnspentListWithBalance struct {
    UnspentList []Unspent
    BalanceList []Balance
}

func (l *Ledger) buildUnspentStorageKey(addr []byte, assets []byte, txHash []byte, outputIdx uint64) (key []byte){
    //prefix + addr + assets hash + transaction hash + output index
    key = []byte(StoragePrefixes.Unspent.String())
    key = append(key, addr...)
    key = append(key, assets...)
    key = append(key, txHash...)
    outputIdxBytes := make([]byte, 8, 8)
    binary.BigEndian.PutUint64(outputIdxBytes, outputIdx)

    key = append(key, outputIdxBytes...)
    key = append(key, )
    return
}

func (l *Ledger) buildBalanceKey(address []byte, assets Assets) (addressBalanceKey []byte, assetsBalanceKey []byte) {
    baseKey := []byte(StoragePrefixes.Balance.String())

    assetsHash := assets.getUniqueHash()
    addressBalanceKey = append(address, assetsHash...)
    addressBalanceKey = append(baseKey, addressBalanceKey...)

    assetsBalanceKey = append(assetsHash, address...)
    assetsBalanceKey = append(baseKey, assetsBalanceKey...)

    return
}

func (l *Ledger) buildUnspentQueryPrefix(addr []byte, assets []byte) (prefix []byte){
    prefix = []byte(StoragePrefixes.Unspent.String())
    prefix = append(prefix, addr...)
    prefix = append(prefix, assets...)
    return
}

func (l *Ledger) getUnspent(key []byte) (unspent Unspent, err error) {
    kv, err := l.Storage.Get(key)
    if err != nil {
        return
    }

    if !kv.Exists {
        err = errors.New("no such unspent")
        return
    }

    err = json.Unmarshal(kv.Data, &unspent)
    return
}

func (l *Ledger) getUnspentListFromTransaction(tx Transaction) (list []Unspent, amount uint64, err error) {
    amount = 0
    for _, ref := range tx.Input {
        key := l.buildUnspentStorageKey(tx.Seal.SignerPublicKey, tx.Assets.getUniqueHash(), ref.Transaction, ref.OutputIndex)
        unspent, dbErr := l.getUnspent(key)
        if dbErr != nil {
            err = errors.New("get Unspent failed: " + dbErr.Error())
            break
        }

        amount += unspent.Value
        list = append(list, unspent)
    }

    return
}

func (l *Ledger) deleteUnspent(tx Transaction) (err error) {
    var keyForDel [][]byte
    for _, ref := range tx.Input {
        key := l.buildUnspentStorageKey(tx.Seal.SignerPublicKey, tx.Assets.getUniqueHash(), ref.Transaction, ref.OutputIndex)
        keyForDel = append(keyForDel, key)
    }

    err = l.Storage.BatchDelete(keyForDel)
    return
}

func (l *Ledger) storeBalance(key []byte, change uint64, isIncrease bool) (amount uint64, err error) {
    bKV, err :=l.Storage.Get(key)
    if err != nil || !bKV.Exists {
        if !isIncrease {
            err = errors.New("invalid address")
            return
        }
        amount = change
    } else {
        current := binary.BigEndian.Uint64(bKV.Data)
        if isIncrease {
            amount = current + change
        } else {
            if current < change {
                err = errors.New("reduce must <= current")
                return
            }

            amount = current - change
        }
    }

    varBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(varBytes, amount)
    kv := kvDatabase.KVItem{
        Key:    key,
        Data:   varBytes,
    }

    err = l.Storage.Put(kv)

    return
}

func (l *Ledger) updateBalance(address []byte, assets Assets, change uint64, isIncrease bool) (amount uint64, err error) {
    addressKey, assetsKey := l.buildBalanceKey(address, assets)
    amount, err = l.storeBalance(addressKey, change, isIncrease)
    if err != nil {
        return
    }

    _, err = l.storeBalance(assetsKey, change, isIncrease)
    if err != nil {
        return
    }

    return

}

type balanceDataInTx struct {
    address     []byte
    assets      Assets
    val         uint64
    isIncrease  bool
}
func (l *Ledger) saveUnspent(localAssets Assets, tx Transaction, in []Unspent) (list UnspentListWithBalance, err error) {
    var unspentList []kvDatabase.KVItem

    balanceIncreaseList := map[string] *balanceDataInTx{}
    assetsHash := localAssets.getUniqueHash()

    for idx, output := range tx.Output {
        key := l.buildUnspentStorageKey(output.To, tx.Assets.getUniqueHash(), tx.Seal.Hash, uint64(idx))

        u := Unspent{
            Owner:          output.To,
            AssetsHash:     assetsHash,
            Transaction:    tx.Seal.Hash,
            OutputIndex:    uint64(idx),
            Value:          output.Value,
        }
        data, _ := json.Marshal(u)

        unspentList = append(unspentList, kvDatabase.KVItem{
            Key: key,
            Data: data,
        })

        bKey := string(output.To) + string(assetsHash)
        if _, exist := balanceIncreaseList[bKey]; exist {
            balanceIncreaseList[bKey].val += output.Value
        } else {
            balanceIncreaseList[bKey] = &balanceDataInTx{
                address: output.To,
                assets: localAssets,
                val: output.Value,
                isIncrease: true,
            }
        }
    }

    err = l.Storage.BatchPut(unspentList)
    if err != nil {
        return
    }

    err = l.deleteUnspent(tx)
    if err != nil {
        return
    }

    balanceReduceList := map[string] *balanceDataInTx{}
    for _, i := range in {
        bKey := string(i.Owner) + string(i.AssetsHash)
        if _, exist := balanceReduceList[bKey]; exist {
            balanceReduceList[bKey].val += i.Value
        } else {
            balanceReduceList[bKey] = &balanceDataInTx{
                address: i.Owner,
                assets: localAssets,
                val: i.Value,
                isIncrease: false,
            }
        }
    }

    var balanceList []Balance
    for _, b := range balanceIncreaseList {
        amount, err := l.updateBalance(b.address, b.assets, b.val, true)
        if err != nil {
            continue
        }
        balanceList = append(balanceList, Balance {
            Address: b.address,
            Assets:  localAssets,
            Amount:  amount,
        })
    }

    for _, b := range balanceReduceList {
        amount, err := l.updateBalance(b.address, b.assets, b.val, false)
        if err != nil {
            continue
        }
        balanceList = append(balanceList, Balance{
            Address: b.address,
            Assets:  localAssets,
            Amount:  amount,
        })
    }

    list.UnspentList = in
    list.BalanceList = balanceList
    return
}

func (l *Ledger) saveUnspentInsideIssueAssetsTransaction(tx Transaction) (balance Balance, err error) {
    u := Unspent{
        Owner:          tx.Assets.IssuedSeal.SignerPublicKey,
        AssetsHash:     tx.Assets.getUniqueHash(),
        Transaction:    tx.Seal.Hash,
        OutputIndex:    uint64(0),
        Value:          tx.Assets.Supply,
    }

    key := l.buildUnspentStorageKey(tx.Assets.IssuedSeal.SignerPublicKey, tx.Assets.getUniqueHash(), tx.Seal.Hash, uint64(0))

    data, _ := json.Marshal(u)
    err = l.Storage.Put(kvDatabase.KVItem{
        Key: key,
        Data: data,
    })

    if err != nil {
        return
    }

    amount, err := l.updateBalance(u.Owner, tx.Assets, u.Value, true)

    balance = Balance{
        Address: u.Owner,
        Assets:  tx.Assets,
        Amount:  amount,
    }
    return
}
