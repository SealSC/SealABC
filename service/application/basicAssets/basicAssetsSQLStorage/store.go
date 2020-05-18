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

package basicAssetsSQLStorage

import (
    "SealABC/log"
    "SealABC/service/application/basicAssets/basicAssetsLedger"
    "SealABC/service/application/basicAssets/basicAssetsTables"
    "encoding/hex"
)

func (s *Storage) StoreAssets(tx basicAssetsLedger.TransactionWithBlockInfo) (err error) {
    assetsRows := basicAssetsTables.AssetsList.NewRows().(basicAssetsTables.AssetsListRows)
    assetsRows.InsertAssets(tx)
    _, err = s.Driver.Insert(&assetsRows, true)
    if err != nil {
        log.Log.Error("insert assets to sql database failed: ", err.Error())
    }

    transfersRows := basicAssetsTables.Transfers.NewRows().(basicAssetsTables.TransfersRows)
    transfersRows.InsertTransferInsideIssueTransaction(tx)
    _, err = s.Driver.Insert(&transfersRows, true)
    if err != nil {
        log.Log.Error("insert transfer to sql database failed: ", err.Error())
    }

    issueToAddr := hex.EncodeToString(tx.Assets.MetaSeal.SignerPublicKey)

    addressRecordRows := basicAssetsTables.AddressRecord.NewRows().(basicAssetsTables.AddressRecordRows)
    addressRecordRows.InsertAddress(tx, issueToAddr, basicAssetsTables.AddressRoles.Issuer)
    _, err = s.Driver.Insert(&addressRecordRows, true)
    if err != nil {
        log.Log.Error("insert address record to sql database failed: ", err.Error())
    }

    addressListRows := basicAssetsTables.AddressList.NewRows().(basicAssetsTables.AddressListRows)
    addressListRows.InsertAddress(tx, issueToAddr)
    _, err = s.Driver.Insert(&addressListRows, true)
    if err != nil {
        log.Log.Error("insert address to sql database failed: ", err.Error())
    }

    return
}

func (s *Storage) StoreUnspent(tx basicAssetsLedger.TransactionWithBlockInfo, inputUnspent []basicAssetsLedger.Unspent) (err error) {
    transfersRows := basicAssetsTables.Transfers.NewRows().(basicAssetsTables.TransfersRows)
    transfersRows.InsertTransfer(tx, inputUnspent)
    _, err = s.Driver.Insert(&transfersRows, true)
    if err != nil {
        log.Log.Error("insert transfer to sql database failed: ", err.Error())
    }

    addressRecordRows := basicAssetsTables.AddressRecord.NewRows().(basicAssetsTables.AddressRecordRows)
    addressRecordRows.InsertAddressesInTransfer(tx, inputUnspent)
    _, err = s.Driver.Insert(&addressRecordRows, true)
    if err != nil {
        log.Log.Error("insert address record to sql database failed: ", err.Error())
    }

    addressListRows := basicAssetsTables.AddressList.NewRows().(basicAssetsTables.AddressListRows)
    addrCache := map[string] bool {}
    for _, out := range tx.Output {
        outAddr := hex.EncodeToString(out.To)
        if _, exists := addrCache[outAddr]; exists{
            continue
        }

        addressListRows.InsertAddress(tx, outAddr)
    }

    for _, in := range inputUnspent {
        inAddr := hex.EncodeToString(in.Owner)
        if _, exists := addrCache[inAddr]; exists{
            continue
        }

        addressListRows.InsertAddress(tx, inAddr)
    }
    _, err = s.Driver.Insert(&addressListRows, true)
    if err != nil {
        log.Log.Error("insert address to sql database failed: ", err.Error())
    }

    return
}

func (s *Storage) StoreBalance(height uint64, tm int64, balanceList []basicAssetsLedger.Balance) (err error) {
    rows := basicAssetsTables.Balance.NewRows().(basicAssetsTables.BalanceRows)
    rows.InsertBalances(height, tm, balanceList)

    _, err = s.Driver.Replace(&rows)
    if err != nil {
        log.Log.Error("insert balance failed: ", err.Error())
    }
    return
}
