# DBTag

it a db columns to go structure project.


## Features
- support mysql , postgres
- special for `xorm` and `gorm`
- support any tags
- generate a .bath/.sh script for next run

## Install

> go get -u -v github.com/athanxx/dbtag

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

