package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	Version = "1.0"
)

func init() {
	caddy.RegisterModule(Database{})
	httpcaddyfile.RegisterHandlerDirective("database", parseCaddyfile)
}

// Middleware implements an HTTP handler that writes the
// uploaded file  to a file on the disk.
type Database struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`

	ctx    caddy.Context
	logger *zap.Logger
	DB     *gorm.DB
}

// CaddyModule returns the Caddy module information.
func (Database) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.database",
		New: func() caddy.Module { return new(Database) },
	}
}

// Provision implements caddy.Provisioner.
func (u *Database) Provision(ctx caddy.Context) error {
	u.ctx = ctx
	u.logger = ctx.Logger(u)

	if u.Driver == "" {
		u.logger.Warn("Provision",
			zap.String("msg", "no FileFieldName specified (file_field_name), using the default one 'myFile'"),
		)
		u.Driver = "mysql"
	}

	if len(u.DSN) < 1 {
		return errors.New("dsn nil")
	}
	if u.Driver == "sqlite" {
		if c, err := gorm.Open(sqlite.Open(u.DSN), &gorm.Config{}); err != nil {
			return fmt.Errorf("open sqlite file error:%s", err.Error())
		} else {
			u.DB = c
		}
	} else {
		if c, err := gorm.Open(mysql.Open(u.DSN), &gorm.Config{}); err != nil {
			return fmt.Errorf("connect mysql error:%s", err.Error())
		} else {
			u.DB = c
		}
	}

	return nil
}

// Validate implements caddy.Validator.
func (u *Database) Validate() error {
	// TODO: Do I need this func
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (u Database) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	u.logger.Info("ServerHTTP", zap.String("Access", r.URL.Path), zap.String("method", r.Method))
	response := map[string]interface{}{}
	req, err := parseRequest(r)
	if err != nil {
		response["code"] = -1
		response["message"] = fmt.Sprintf("parse request error:%s", err.Error())
	} else {
		data, err := req.UseDB(u.DB).Handle()
		if err != nil {
			response["code"] = -2
			response["message"] = err.Error()
			response["data"] = nil
		} else {
			response["code"] = 0
			response["message"] = "success"
			response["data"] = data
		}
	}

	bytes, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(bytes)

	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (u *Database) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {

	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "driver":
				if !d.Args(&u.Driver) {
					return d.ArgErr()
				}
			case "dsn":
				if !d.Args(&u.DSN) {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized servers option '%s'", d.Val())
			}
		}
	}
	return nil
}

// parseCaddyfile parses the upload directive. It enables the upload
// of a file:
//
//	upload {
//	    dest_dir          <destination directory>
//	    max_filesize      <size>
//	    response_template [<path to a response template>]
//	}
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var u Database
	err := u.UnmarshalCaddyfile(h.Dispenser)
	return u, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Database)(nil)
	_ caddy.Validator             = (*Database)(nil)
	_ caddyhttp.MiddlewareHandler = (*Database)(nil)
	_ caddyfile.Unmarshaler       = (*Database)(nil)
)
