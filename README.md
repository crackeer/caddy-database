## Requirement
- [Go Installed](https://golang.org/doc/install)

## Development

```sh
xcaddy run --config ./caddy.json
```

## Install
refer to [Extending Caddy](https://caddyserver.com/docs/extending-caddy)
1. **Install [xcaddy](https://github.com/caddyserver/xcaddy)**

```sh
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
```

2. **Build A New Caddy Binary**

```sh
xcaddy build master --with github.com/crackeer/caddy-database
```

3. **copy new template.html**

here is the [template.html](https://github.com/crackeer/caddy-upload2dir/blob/main/template.html)

## Example:caddy.json
apps.http.servers下的一个配置
```json
{
    "admin": {
        "disabled": false,
        "listen": "0.0.0.0:2019",
        "enforce_origin": false,
        "origins": [
            ""
        ]
    },
    "apps": {
        "http": {
            "servers": {
                "static": {
                    "idle_timeout": 30000000000,
                    "listen": [
                        "0.0.0.0:8080"
                    ],
                    "max_header_bytes": 10240000,
                    "read_header_timeout": 10000000000,
                    "routes": [
                        {
                            "match": [
                                {
                                    "header": {
                                        "proxy": [
                                            "database"
                                        ]
                                    }
                                }
                            ],
                            "handle": [
                                {
                                    "handler": "database",
                                    "driver": "mysql",
                                    "dsn": "USER:PASSWORD@tcp(HOST:PORT)/DATABASE?charset=utf8mb4&parseTime=True&loc=Local"
                                }
                            ],
                            "terminal": true
                        }
                    ]
                }
            }
        }
    }
}
```

#### about user_config
`token`:`username`:`create_dir`/`put_file`/`delete_file`
There are three actions you can config in user_config
- create_dir
- put_file
- delete_file

## what new filer_server page looks like?
- Add create directory in current directory、upload file to current directory、delete file or empty directory
[![pppwtDU.png](https://s1.ax1x.com/2023/02/26/pppwtDU.png)](https://imgse.com/i/pppwtDU)

**attention!!!** 
you have to set your token which represents your identity that make you have access to the three actions, you can click `set token` at the bottom of the page.
[![pppwBCR.png](https://s1.ax1x.com/2023/02/26/pppwBCR.png)](https://imgse.com/i/pppwBCR)


