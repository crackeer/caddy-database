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

	actionCreate string = "create"
	actionInsert string = "insert"
	actionUpdate string = "update"
	actionDelete string = "delete"

	actionShowTables string = "show_tables"

	actionDrop string = "drop"
)

// Request
type Request struct {
	Action string
	Table  string
	Body   []byte

	DB *gorm.DB
}

// SelectBody
type SelectBody struct {
	Query    map[string]interface{} `json:"query"`
	Fields   []string               `json:"fields"`
	OrderBy  []string               `json:"order_by"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
	Distinct []string               `json:"distinct"`
}

// UpdateBody ...
type UpdateBody struct {
	Query  map[string]interface{} `json:"query"`
	Update map[string]interface{} `json:"update"`
	Limit  int                    `json:"limit"`
}

func parseRequest(r *http.Request) (*Request, error) {
	path := strings.Trim(r.URL.RequestURI(), "/")
	if len(path) < 1 {
		return nil, errors.New("nil path")
	}
	parts := strings.Split(path, "/")
	action := parts[0]

	if len(parts) < 2 {
		return nil, errors.New("nil table")
	}

	bytes, _ := ioutil.ReadAll(r.Body)

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
func (req *Request) UseDB(db *gorm.DB) *Request {
	req.DB = db
	return req
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
	if req.Action == actionInsert {
		return req.Insert()
	}
	if req.Action == actionUpdate {
		return req.Update()
	}
	if req.Action == actionDelete {
		return req.Delete()
	}
	if req.Action == actionCreate {
		return req.Create()
	}

	return nil, nil
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

	list := []string{}

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
	data := []map[string]interface{}{}

	if err := req.DB.Raw("show tables").Scan(&data).Error; err != nil {
		return nil, err
	}
	retData := []string{}
	for _, value := range data {
		for _, value := range value {
			retData = append(retData, value.(string))
		}
	}

	return retData, nil
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
	where, params := BuildQuery(updateBody.Query)
	db := req.DB.Table(req.Table).Where(where, params...).Updates(updateBody.Update)
	if db.Error != nil {
		return nil, db.Error
	}

	return db.RowsAffected, nil
}

// Delete
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Delete() (interface{}, error) {
	updateBody := &UpdateBody{}
	if err := req.decodeBody(updateBody); err != nil {
		return nil, fmt.Errorf("decode update body error:%s", err.Error())
	}

	where, params := BuildQuery(updateBody.Query)
	sql := fmt.Sprintf("delete from %s where %s", req.Table, where)
	return req.DB.Exec(sql, params...).RowsAffected, nil
}

// Create
//
//	@receiver req
//	@return interface{}
//	@return error
func (req *Request) Create() (interface{}, error) {
	if err := req.DB.Exec(string(req.Body)).Error; err != nil {
		return nil, err
	}
	return nil, nil
}
