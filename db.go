package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

const (
	actionSelect          string = "select"
	actionCount           string = "count"
	actionDistinct        string = "distinct"
	actionShowCreateTable string = "show_create_table"
	actionDesc            string = "desc"

	actionInsert string = "insert"
	actionUpdate string = "update"
	actionDelete string = "delete"

	actionShowTables string = "show_tables"

	actionDrop string = "drop"
	actionExec string = "exec"
)

// Request
type Request struct {
	Action string
	Table  string
	Body   []byte

	DB     *gorm.DB
	driver string
}

// SelectBody
type SelectBody struct {
	Query    map[string]interface{} `json:"query"`
	Fields   []string               `json:"fields"`
	OrderBy  string                 `json:"order_by"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
	Distinct []string               `json:"distinct"`
}

// UpdateBody
type UpdateBody struct {
	Query  map[string]interface{} `json:"query"`
	Update map[string]interface{} `json:"update"`
}

func parseRequest(r *http.Request) (*Request, error) {
	path := strings.Trim(r.URL.RequestURI(), "/")
	if len(path) < 1 {
		return nil, errors.New("nil path")
	}
	parts := strings.Split(path, "/")
	action := parts[0]
	bytes, _ := ioutil.ReadAll(r.Body)

	if len(parts) == 1 {
		if action == actionExec || action == actionShowTables {
			return &Request{
				Action: action,
				Body:   bytes,
			}, nil
		}
		return nil, errors.New("table nil")
	}

	return &Request{
		Action: action,
		Table:  parts[1],
		Body:   bytes,
	}, nil
}

// UseDB
//
//	@receiver req
//	@param db
//	@return *Request
func (req *Request) UseDB(db *gorm.DB, driver string) *Request {
	req.DB = db
	req.driver = driver
	return req
}

// IsSQLite
//
//	@receiver req
//	@return bool
func (req *Request) IsSQLite() bool {
	return req.driver == "sqlite"
}

// Handle
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Handle() (interface{}, error) {
	if req.Action == actionSelect {
		return req.Select()
	}

	if req.Action == actionCount {
		return req.Count()
	}

	if req.Action == actionDistinct {
		return req.Distinct()
	}

	if req.Action == actionShowCreateTable {
		return req.ShowCreateTable()
	}

	if req.Action == actionDesc {
		return req.Desc()
	}

	if req.Action == actionShowTables {
		return req.ShowTables()
	}

	if req.Action == actionExec {
		return req.Exec()
	}

	if req.Action == actionInsert {
		return req.Insert()
	}

	if req.Action == actionUpdate {
		return req.Update()
	}

	if req.Action == actionDelete {
		return req.Delete()
	}

	if req.Action == actionDrop {
		return req.Drop()
	}

	return nil, errors.New("no action match")
}

func (req *Request) decodeBody(dest interface{}) error {
	return json.Unmarshal(req.Body, dest)
}

// Select
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Select() (interface{}, error) {
	selectBody := &SelectBody{}
	if err := req.decodeBody(selectBody); err != nil {
		return nil, fmt.Errorf("decode select body error:%s", err.Error())
	}

	db := req.DB.Table(req.Table)
	if len(selectBody.Fields) > 0 {
		db = db.Select(selectBody.Fields)
	}

	if len(selectBody.Query) > 0 {
		sql, params := BuildQuery(selectBody.Query)
		db = db.Where(sql, params...)
	}

	if len(selectBody.OrderBy) > 0 {
		db = db.Order(selectBody.OrderBy)
	}

	if selectBody.Offset > 0 {
		db = db.Offset(selectBody.Offset)
	}
	if selectBody.Limit > 0 {
		db = db.Limit(selectBody.Limit)
	}

	list := []map[string]interface{}{}

	if err := db.Find(&list).Error; err != nil {
		return nil, fmt.Errorf("select error:%s", err.Error())
	}

	return list, nil
}

// Select
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Count() (interface{}, error) {
	selectBody := &SelectBody{}
	if err := req.decodeBody(selectBody); err != nil {
		return nil, fmt.Errorf("decode select body error:%s", err.Error())
	}

	db := req.DB.Table(req.Table)
	if len(selectBody.Query) > 0 {
		sql, params := BuildQuery(selectBody.Query)
		db = db.Where(sql, params...)
	}

	var count int64
	if len(selectBody.Distinct) > 0 {
		db = db.Distinct(selectBody.Distinct[0])
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("count error:%s", err.Error())
	}

	return count, nil
}

// Distinct
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Distinct() (interface{}, error) {
	selectBody := &SelectBody{}
	if err := req.decodeBody(selectBody); err != nil {
		return nil, fmt.Errorf("decode select body error:%s", err.Error())
	}

	if len(selectBody.Distinct) < 1 {
		return nil, fmt.Errorf("distinct colum nil")
	}

	db := req.DB.Table(req.Table)
	if len(selectBody.Query) > 0 {
		sql, params := BuildQuery(selectBody.Query)
		db = db.Where(sql, params...)
	}

	db = db.Distinct(selectBody.Distinct[0])

	list := []interface{}{}

	if err := db.Pluck(selectBody.Distinct[0], &list).Error; err != nil {
		return nil, fmt.Errorf("count error:%s", err.Error())
	}

	return list, nil
}

// Show
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) ShowCreateTable() (interface{}, error) {

	data := map[string]interface{}{}
	if req.IsSQLite() {
		req.DB.Table("sqlite_master").Where(map[string]interface{}{
			"type": "table",
			"name": req.Table,
		}).Scan(&data)
		return data["sql"], nil
	}
	if err := req.DB.Raw("show create table " + req.Table).Scan(&data).Error; err != nil {
		return nil, err
	}
	if value, ok := data["Create Table"]; ok {
		if stringValue, ok := value.(string); ok {
			return stringValue, nil
		}
	}

	return "", nil
}

// Desc
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Desc() (interface{}, error) {
	list := []map[string]interface{}{}
	if req.IsSQLite() {
		if err := req.DB.Raw(fmt.Sprintf("PRAGMA TABLE_INFO (%s)", req.Table)).Scan(&list).Error; err != nil {
			return nil, err
		}
		return list, nil
	}

	if err := req.DB.Raw("desc " + req.Table).Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

// ShowTables
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) ShowTables() (interface{}, error) {
	list := []map[string]interface{}{}
	retData := []string{}
	if req.IsSQLite() {
		req.DB.Table("sqlite_master").Where(map[string]interface{}{
			"type": "table",
		}).Find(&list)
		for _, value := range list {
			for key, value := range value {
				if key == "name" {
					retData = append(retData, value.(string))
				}
			}
		}
		return retData, nil
	}
	if err := req.DB.Raw("show tables").Scan(&list).Error; err != nil {
		return nil, err
	}
	for _, value := range list {
		for _, value := range value {
			retData = append(retData, value.(string))
		}
	}

	return retData, nil
}

// Exec
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Exec() (interface{}, error) {
	if err := req.DB.Exec(string(req.Body)).Error; err != nil {
		return nil, err
	}
	return "ok", nil
}

// Insert
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Insert() (interface{}, error) {
	list := []map[string]interface{}{}
	if err := req.decodeBody(&list); err != nil {
		return nil, fmt.Errorf("decode insert body error:%s", err.Error())
	}

	if err := req.DB.Table(req.Table).Create(&list).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

// Update
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Update() (interface{}, error) {
	updateBody := &UpdateBody{}
	if err := req.decodeBody(updateBody); err != nil {
		return nil, fmt.Errorf("decode update body error:%s", err.Error())
	}

	sql, params := BuildQuery(updateBody.Query)

	db := req.DB.Table(req.Table).Where(sql, params...).Updates(updateBody.Update)
	if err := db.Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"affected_rows": db.RowsAffected,
	}, nil
}

func (req *Request) Delete() (interface{}, error) {
	updateBody := &UpdateBody{}
	if err := req.decodeBody(updateBody); err != nil {
		return nil, fmt.Errorf("decode update body error:%s", err.Error())
	}

	where, params := BuildQuery(updateBody.Query)
	sql := fmt.Sprintf("delete from %s where %s", req.Table, where)
	db := req.DB.Exec(sql, params...)
	if err := db.Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"affected_rows": db.RowsAffected,
	}, nil
}

// Drop
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Drop() (interface{}, error) {
	db := req.DB.Exec("drop table " + req.Table)
	if err := db.Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"affected_rows": db.RowsAffected,
	}, nil
}
