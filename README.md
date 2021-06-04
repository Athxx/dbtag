# DBTag

it a db columns to go structure project.


## Features
- support mysql , postgres
- special for `xorm` and `gorm`
- support any tags
- generate a .bath/.sh script for next run

## Install

> go get -u -v github.com/athanxx/dbtag

or you can download binary files for `Windows` and `Linux`

> https://github.com/athanxx/dbtag/releases

## Example 

> dbtag

or use args

> dbtag -db_addr=127.0.0.1:3306 -db_auth=root:123456 -db_name=abc -adapter=postgres -dir=./model -fn=y   

it will ask what tables you want to generate, supported any tag what you want. 

supported `xorm` and `gorm` formats to generate template structure.

## Args

`-db_addr` db host and port `127.0.0.1:3306`

`-db_auth` username and passcode `root:123456`

`-db_name` database name `abc`

`-adapter` support `mysql` and `postgres` 

`-dir` generate directory path, pack name same with directory name `./abc/model`

`-tag` any tags, special for `xorm` or `gorm` template, `xorm,gorm,json,db`

`-t` table list, use ',' to split it `user_info,user_list`

`-fn` it will create a function Table() to get table name `y` or others 

## Demo

```go
package model
import (
	"time"
)
/*
CREATE TABLE `a` (
  `a` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '1',
  `b` date NOT NULL DEFAULT '2000-01-01' COMMENT '2',
  `c` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '3',
  `d` json NOT NULL COMMENT '4',
  `e` varchar(100) NOT NULL DEFAULT 'abc' COMMENT '5',
  PRIMARY KEY (`a`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
*/

type A struct{
	A int       `db:"a" json:"a" yaml:"a" xorm:"a int(10) unsigned pk not null default(0) comment(1)" gorm:"primary_key;column:a;type:int(10) unsigned;not null;default:0;comment:1" toml:"a"`
	B string    `db:"b" json:"b" yaml:"b" xorm:"b date not null default(2000-01-01) comment(2)" gorm:"column:b;type:date;not null;default:2000-01-01;comment:2" toml:"b"`
	C time.Time `db:"c" json:"c" yaml:"c" xorm:"c datetime not null default(CURRENT_TIMESTAMP) comment(3)" gorm:"column:c;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:3" toml:"c"`
	D string    `db:"d" json:"d" yaml:"d" xorm:"d json not null comment(4)" gorm:"column:d;type:json;not null;comment:4" toml:"d"`
	E string    `db:"e" json:"e" yaml:"e" xorm:"e varchar(100) not null default(abc) comment(5)" gorm:"column:e;type:varchar(100);not null;default:abc;comment:5" toml:"e"`
}

func (t *A) Table() string {
	return "a"
}
```

