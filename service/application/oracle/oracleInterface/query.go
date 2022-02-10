package oracleInterface

import (
	"github.com/SealSC/SealABC/metadata/blockchainRequest"
	"errors"
	"strings"
	"github.com/SealSC/SealABC/service/system/blockchain/chainTables"
	"github.com/SealSC/SealABC/storage/db/dbInterface/simpleSQLDatabase"
	"fmt"
	"encoding/hex"
	"encoding/json"
)

func (o *OracleApplication) requestDispatch(req blockchainRequest.Entity) (resp interface{}, err error) {
	switch req.QueryString {
	case "":
		return o.querySaveStatus(req)
	case "/cron":
		return o.cronOption(req)
	}
	return nil, errors.New("req.queryString is not supported")
}

func (o *OracleApplication) querySaveStatus(req blockchainRequest.Entity) (resp interface{}, err error) {
	if strings.TrimSpace(req.RequestAction) == "" {
		return nil, errors.New("action not found")
	}
	//cache   DB
	//yes     no,    Saving
	//no      no,    Notfound
	//no      yes,   Saved
	//yes     yes,   -
	_, cached, err := o.PoolGet(req)
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
	rows, err := o.blockchainDriver.Query(chainTables.RequestRow{}, sqlStr, []interface{}{hex.EncodeToString(req.Seal.Hash)})
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

func (o *OracleApplication) cronOption(req blockchainRequest.Entity) (resp CronResponse, err error) {
	cronReq := CronRequest{}
	if err = json.Unmarshal(req.Data, &cronReq); err != nil {
		return
	}
	resp = CronResponse{
		ID:   cronReq.ID,
		Path: cronReq.Path,
	}

	switch cronReq.Action {
	case ActionInsert: //insert
		resp.ID, err = o.addSchedule(cronReq.Path, req.RequestAction)
		return
	case ActionDelete: //delete
		o.cr.Remove(cronReq.ID)
		return
	case ActionUpdate: //update
		o.cr.Remove(cronReq.ID)
		resp.ID, err = o.addSchedule(cronReq.Path, req.RequestAction)
		return
	case ActionQueryOne: //query
		resp.Path = o.cr.Query(cronReq.ID)
	case ActionQueryList: //query
		resp.M = o.cr.crRunning
	default:
		return CronResponse{}, errors.New("cron.action is not supported")
	}
	return
}
