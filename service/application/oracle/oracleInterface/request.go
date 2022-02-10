package oracleInterface

import "encoding/json"

const (
	ActionInsert    = 1
	ActionDelete    = 2
	ActionUpdate    = 3
	ActionQueryOne  = 4
	ActionQueryList = 5
)

type CronRequest struct {
	//添加
	//查询(通过ID查询表达式)
	//删除
	//修改(只能修改)

	//定时器的ID
	//cron表达式
	ID     int
	Path   string
	Action int //增删改查
}

type CronResponse struct {
	ID   int
	Path string
	M    map[int]string
}

func (c CronResponse) MarshalJSON() ([]byte, error) {
	if c.M != nil {
		return json.Marshal(c.M)
	}
	return json.Marshal(map[string]interface{}{"id": c.ID, "path": c.Path})
}
