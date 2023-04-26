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
	actionSelect string = "select"
	actionUpdate string = "update"
	actionDelete string = "delete"
	actionDrop   string = "drop"
	actionExec   string = "exec"
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
	Query   map[string]interface{} `json:"query"`
	Fields  []string               `json:"fields"`
	OrderBy []string               `json:"order_by"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
}

func parseRequest(r *http.Request) (*Request, error) {
	path := strings.Trim(r.URL.RequestURI(), "/")
	if len(path) < 1 {
		return nil, errors.New("nil path")
	}
	parts := strings.Split(path, "/")
	action := parts[0]

	if action != actionExec && len(parts) < 2 {
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
