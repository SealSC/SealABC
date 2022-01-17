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
	c                http.Client
	functions        map[string]Action
	reqPool          map[string]blockchainRequest.Entity
	blockchainDriver simpleSQLDatabase.IDriver
	reqMaxCache      int
}
type Action interface {
	Name() string
	VerifyReq(req []byte) error
	FormatResult(req blockchainRequest.Entity) (applicationResult.Entity, error)
}

var defVerifyConnection = func(state tls.ConnectionState) error { return errors.New("empty Verify Connection") }

func NewOracleApplication(reqMaxCache int, pullTimeOut time.Duration, blockchainDriver simpleSQLDatabase.IDriver) *OracleApplication {
	return &OracleApplication{
		blockchainDriver: blockchainDriver,
		reqPool:          make(map[string]blockchainRequest.Entity, reqMaxCache),
		reqMaxCache:      reqMaxCache,
		functions:        map[string]Action{},
		c: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false, VerifyConnection: defVerifyConnection},
			},
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       pullTimeOut,
		},
	}
}

func (o *OracleApplication) RegFunction(a Action) {
	o.functions[a.Name()] = a
}

func blockchainRequestEntityKey(e blockchainRequest.Entity) string {
	return e.RequestAction + "+" + e.Seal.HexHash()
}
func (o *OracleApplication) PoolAdd(e blockchainRequest.Entity) {
	o.Lock()
	defer o.Unlock()
	o.reqPool[blockchainRequestEntityKey(e)] = e
}
func (o *OracleApplication) PoolGet(e blockchainRequest.Entity) (blockchainRequest.Entity, bool) {
	o.Lock()
	defer o.Unlock()
	entity, ok := o.reqPool[blockchainRequestEntityKey(e)]
	return entity, ok
}
func (o *OracleApplication) PoolDelete(e blockchainRequest.Entity) {
	o.Lock()
	defer o.Unlock()
	delete(o.reqPool, blockchainRequestEntityKey(e))
}
func (o *OracleApplication) PoolGetAll() (entity []blockchainRequest.Entity) {
	o.Lock()
	defer o.Unlock()
	for s := range o.reqPool {
		entity = append(entity, o.reqPool[s])
	}
	return
}

func (o *OracleApplication) Name() (name string) {
	return "Oracle"
}

func (o *OracleApplication) reqValidate(req blockchainRequest.Entity) error {
	if req.RequestApplication != o.Name() {
		return errors.New("not found RequestApplication")
	}
	if len(req.Seal.Hash) == 0 {
		return errors.New("Seal.Hash is empty")
	}
	return nil
}

func (o *OracleApplication) PushClientRequest(req blockchainRequest.Entity) (result interface{}, err error) {
	err = o.reqValidate(req)
	if err != nil {
		return
	}
	if len(o.reqPool) >= o.reqMaxCache {
		return nil, errors.New("is running")
	}
	o.PoolAdd(req)
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
	_, cached := o.PoolGet(str)
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

func (o *OracleApplication) PreExecute(req blockchainRequest.Entity, header block.Entity) (_ []byte, err error) {
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
	err = a.VerifyReq(req.Data)
	return
}

func (o *OracleApplication) Execute(req blockchainRequest.Entity, header block.Entity, actIndex uint32) (result applicationResult.Entity, err error) {
	o.PoolDelete(req)
	a := o.functions[req.RequestAction]
	if a == nil {
		err = errors.New("action not found")
		return
	}
	if err = a.VerifyReq(req.Data); err != nil {
		return
	}

	result, err = a.FormatResult(req)
	if err != nil {
		return
	}
	if result.Seal != nil {
		header.EntityData.Body.RequestsCount += 1
		data := blockchainRequest.Entity{}
		data.Data, _ = json.Marshal(result.Data)
		data.Seal = *result.Seal
		header.EntityData.Body.Requests = append(header.EntityData.Body.Requests, data)
	}
	return
}

func (o *OracleApplication) Cancel(req blockchainRequest.Entity) (err error) {
	return nil
}

func (o *OracleApplication) RequestsForBlock(_ block.Entity) (entity []blockchainRequest.Entity, cnt uint32) {
	//new view
	entity = o.PoolGetAll()
	cnt = uint32(len(entity))
	return
}

func (o *OracleApplication) ApplicationInternalCall(src string, callData []byte) (ret interface{}, err error) {
	return
}

func (o *OracleApplication) Information() (info service.BasicInformation) {
	//http api
	return
}

func (o *OracleApplication) SetChainInterface(ci chainStructure.IChainInterface) {
	//new server
	return
}

func (o *OracleApplication) UnpackingActionsAsRequests(req blockchainRequest.Entity) (reqList []blockchainRequest.Entity, err error) {
	//add block
	return
}

func (o *OracleApplication) GetActionAsRequest(req blockchainRequest.Entity) (ret blockchainRequest.Entity) {
	//add block
	return req
}
