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

package oracleInterface

import (
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/metadata/applicationResult"
	"github.com/SealSC/SealABC/service"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"errors"
	"sync"
	"net/http"
	"encoding/json"
	"time"
	"crypto/tls"
	"encoding/hex"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"github.com/SealSC/SealABC/service/system/blockchain/chainTables"
	"fmt"
	"strings"
	"github.com/robfig/cron/v3"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/storage/db/dbInterface/kvDatabase"
)

type queryResult struct {
	Status    int         `json:"status"`
	StatusMSG string      `json:"statusMSG"`
	Data      interface{} `json:"data"`
}

func queryResultSaving() *queryResult {
	return &queryResult{
		Status:    3,
		StatusMSG: "Saving",
		Data:      nil,
	}
}
func queryResultSaved(data interface{}) *queryResult {
	return &queryResult{
		Status:    2,
		StatusMSG: "Saved",
		Data:      data,
	}
}
func queryResultNotfound() *queryResult {
	return &queryResult{
		Status:    4,
		StatusMSG: "Notfound",
		Data:      nil,
	}
}

type OracleApplication struct {
	sync.Mutex
	cr               *cron.Cron
	pullTimeOut      time.Duration
	functions        map[string]Action
	reqPool          *SortMap
	blockchainDriver simpleSQLDatabase.IDriver
}

func (o *OracleApplication) removeSchedule(ids ...int) {
	for _, id := range ids {
		o.cr.Remove(cron.EntryID(id))
	}
}

func (o *OracleApplication) addSchedule(cronPath string, name string) (int, error) {
	s := o.functions[name]
	z, ok := s.(ActionRemoteAutoPuller)
	if !ok {
		return -1, nil
	}
	var f = func() {
		url, contentType := z.UrlContentType()
		client := http.Client{
			Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false, VerifyConnection: z.VerifyRemoteCA}},
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       o.pullTimeOut,
		}
		post, err := client.Post(url, contentType, nil)
		if err != nil {
			log.Log.Errorln(err)
			return
		}
		e, err := z.VerifyRemoteResp(post)
		if err != nil {
			log.Log.Errorln(err)
			return
		}
		e.RequestApplication = "Oracle"
		e.RequestAction = z.Name()
		e.Seal.HexHash()
		err = o.PoolAdd(e, true)
		if err != nil {
			log.Log.Errorln(err)
		}
	}
	id, err := o.cr.AddFunc(cronPath, f)
	return int(id), err
}

func NewOracleApplication(pullTimeOut time.Duration,
	blockchainDriver simpleSQLDatabase.IDriver,
	kvDB kvDatabase.IDriver) *OracleApplication {
	o := &OracleApplication{
		blockchainDriver: blockchainDriver,
		reqPool:          NewSortMap(kvDB),
		functions:        map[string]Action{},
		cr:               cron.New(cron.WithSeconds()), //Second | Minute | Hour | Dom | Month | Dow | Descriptor
		pullTimeOut:      pullTimeOut,
	}
	o.cr.Start()
	return o
}

func (o *OracleApplication) RegFunction(a Action) error {
	if isBaseAction(a) {
		return errors.New("action is base action, unable to register")
	}

	o.functions[a.Name()] = a
	puller, ok := a.(ActionRemoteAutoPuller)
	if !ok {
		return nil
	}
	paths := puller.CronPaths()
	var ids []int
	for _, pat := range paths {
		id, err := o.addSchedule(pat, puller.Name())
		if err != nil {
			o.removeSchedule(ids...)
			return err
		}
		if id >= 0 {
			ids = append(ids, id)
		}
	}
	return nil
}

func blockchainRequestEntityKey(e blockchainRequest.Entity) string {
	return e.RequestAction + "+" + e.Seal.HexHash()
}
func (o *OracleApplication) PoolAdd(e blockchainRequest.Entity, remote bool) error {
	o.Lock()
	defer o.Unlock()
	if e.IsFromNull() {
		if remote {
			e.FromRemote()
		} else {
			e.FromAPI()
		}
	}
	return o.reqPool.set(blockchainRequestEntityKey(e), e)
}
func (o *OracleApplication) PoolGet(e blockchainRequest.Entity) (blockchainRequest.Entity, bool, error) {
	o.Lock()
	defer o.Unlock()
	entity, ok, err := o.reqPool.get(blockchainRequestEntityKey(e))
	return entity, ok, err
}

func (o *OracleApplication) PoolOne() (blockchainRequest.Entity, bool) {
	o.Lock()
	defer o.Unlock()
	list := o.reqPool.list()
	if len(list) <= 0 {
		return blockchainRequest.Entity{}, false
	}
	return list[0], true
}

func (o *OracleApplication) PoolDelete(e blockchainRequest.Entity) error {
	o.Lock()
	defer o.Unlock()
	return o.reqPool.del(blockchainRequestEntityKey(e))
}

func (o *OracleApplication) Name() (name string) {
	return "Oracle"
}

func (o *OracleApplication) reqValidate(req blockchainRequest.Entity) error {
	if req.IsFromRemote() {
		_, ok, err := o.PoolGet(req)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("not found remote data")
		}
	}
	if req.RequestApplication != o.Name() {
		return errors.New("not found RequestApplication")
	}
	return nil
}

func (o *OracleApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
	err = o.reqValidate(req)
	if err != nil {
		return
	}
	err = o.PoolAdd(req, false)
	return
}

func (o *OracleApplication) Query(req []byte) (interface{}, error) {
	str := blockchainRequest.Entity{}
	err := json.Unmarshal(req, &str)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(str.RequestAction) == "" {
		return nil, errors.New("action not found")
	}
	//cache   DB
	//yes     no,    Saving
	//no      no,    Notfound
	//no      yes,   Saved
	//yes     yes,   -
	_, cached, err := o.PoolGet(str)
	if err != nil {
		return nil, err
	}
	if cached {
		return queryResultSaving(), nil
	}
	if o.blockchainDriver == nil {
		return queryResultNotfound(), nil
	}

	table := chainTables.RequestsTable{}
	tag, err := simpleSQLDatabase.ColumnsFromTag(table, false, nil)
	if err != nil {
		return nil, err
	}
	sqlStr := fmt.Sprintf("SELECT `%v` FROM `%s` WHERE `c_hash`=?", strings.Join(tag, "`,`"), table.Name())
	rows, err := o.blockchainDriver.Query(chainTables.RequestRow{}, sqlStr, []interface{}{hex.EncodeToString(str.Seal.Hash)})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return queryResultNotfound(), nil
	}
	requestsTable, ok := rows[0].(chainTables.RequestRow)
	if !ok {
		return nil, errors.New("result not a RequestRow")
	}
	result := map[string]interface{}{}
	result["hash"] = requestsTable.Hash
	result["time"] = requestsTable.Time
	result["height"] = requestsTable.Height
	payload := map[string]interface{}{}
	err = json.Unmarshal([]byte(requestsTable.Payload), &payload)
	if err != nil {
		return nil, err
	}
	result["payload"] = payload
	return queryResultSaved(result), nil
}

func (o *OracleApplication) VerifyReq(a Action, req blockchainRequest.Entity) (err error) {
	verifier, ok := a.(RequestVerifier)
	if !ok {
		err = errors.New("action not support http request")
	}
	return verifier.VerifyReq(req)
}

func (o *OracleApplication) PreExecute(req blockchainRequest.Entity, _ block.Entity) (_ []byte, err error) {
	//validate trusting
	if err = o.reqValidate(req); err != nil {
		return nil, err
	}
	req.PackedCount = 0
	req.Packed = false

	a := o.functions[req.RequestAction]
	if a == nil {
		err = errors.New("action not found")
		return
	}
	err = o.VerifyReq(a, req)
	return
}

func (o *OracleApplication) Execute(req blockchainRequest.Entity, header block.Entity, _ uint32) (result applicationResult.Entity, err error) {
	err = o.reqValidate(req)
	if err != nil {
		return
	}
	err = o.PoolDelete(req)
	if err != nil {
		return
	}
	a := o.functions[req.RequestAction]
	if a == nil {
		err = errors.New("action not found")
		return
	}
	if err = o.VerifyReq(a, req); err != nil {
		return
	}

	data, err := a.FormatResult(req)
	if err != nil {
		return result, err
	}
	result.Data = data
	result.Seal = &req.Seal

	header.EntityData.Body.RequestsCount += 1
	blockData := blockchainRequest.Entity{}
	blockData.Data, _ = json.Marshal(result.Data)
	blockData.Seal = *result.Seal
	header.EntityData.Body.Requests = append(header.EntityData.Body.Requests, blockData)

	return
}

func (o *OracleApplication) Cancel(blockchainRequest.Entity) (err error) {
	return nil
}

func (o *OracleApplication) RequestsForBlock(_ block.Entity) (entity []blockchainRequest.Entity, cnt uint32) {
	//new view
	one, ok := o.PoolOne()
	if ok {
		cnt = 1
		entity = append(entity, one)
	}
	return
}

func (o *OracleApplication) ApplicationInternalCall(string, []byte) (ret interface{}, err error) {
	return
}

func (o *OracleApplication) Information() (info service.BasicInformation) {
	//http api
	return
}

func (o *OracleApplication) SetChainInterface(chainStructure.IChainInterface) {
	//new server
	return
}

func (o *OracleApplication) UnpackingActionsAsRequests(blockchainRequest.Entity) (reqList []blockchainRequest.Entity, err error) {
	//add block
	return
}

func (o *OracleApplication) GetActionAsRequest(req blockchainRequest.Entity) (ret blockchainRequest.Entity) {
	//add block
	return req
}
