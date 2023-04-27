## Create Table

```curl
curl --location 'localhost:8080/exec' \
--header 'proxy: sqlite' \
--header 'Content-Type: text/plain' \
--data 'CREATE TABLE COMPANY(
   ID INT PRIMARY KEY     NOT NULL,
   NAME           TEXT    NOT NULL,
   AGE            INT     NOT NULL,
   ADDRESS        CHAR(50),
   SALARY         REAL
);'
```


## Insert Data

```curl
curl --location 'localhost:8080/insert/company' \
--header 'proxy: sqlite' \
--header 'Content-Type: application/json' \
--data '[
    {
        "id" : 1,
        "name":"simple",
        "age" : 89,
        "address" : "simple"
    }
]'
```

## Update Data

```curl
curl --location 'localhost:8080/update/company' \
--header 'proxy: sqlite' \
--header 'Content-Type: application/json' \
--data '{
    "query" : {
        "id" : 111
    },
    "update" : {
        "salary" : 122
    }
}'
```