package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

/*

   you can using dbTag to get any tag from database
   support postgresql and mysql
   you can use args or cmd to run it.

*/

var (
	PackName string
	TabList  []string
	Dir      = "./model"
	Tag      = ""
	Adapter  = "mysql"
	TableFn  = false
	DBAuth   string
	DBAddr   string
	DBName   string
	SqlNull  bool
)

type Col struct {
	Field      string `db:"Field"`
	Type       string `db:"Type"`
	Collation  []byte `db:"Collation"`
	Null       []byte `db:"Null"`
	Key        string `db:"Key"`
	Default    []byte `db:"Default"`
	Extra      string `db:"Extra"`
	Privileges string `db:"Privileges"`
	Comment    string `db:"Comment"`
}

func main() {
	base := checkArgs()
	var err error
	//fmt.Println(Adapter, DBAuth+"@tcp("+DBAddr+")/"+DBName+"?charset=utf8mb4")
	db, _ := sqlx.Connect(Adapter, DBAuth+"@tcp("+DBAddr+")/"+DBName+"?charset=utf8mb4")
	if err = db.Ping(); err != nil {
		fmt.Println("Cannot connect db server : " + err.Error())
		return
	}
	// get tables list
	tbs := TabList
	if len(tbs) == 0 {
		if err = db.Select(&tbs, "SHOW TABLES;"); err != nil {
			fmt.Println("Cannot get tables : " + err.Error())
			return
		}
	}

	for _, tbName := range tbs {
		sql := ""
		// get table create table sql
		rows, _ := db.Query("SHOW CREATE TABLE " + tbName)
		for rows.Next() {
			rows.Scan(&sql, &sql)
		}

		reg, _ := regexp.Compile(`(?i)AUTO_INCREMENT=\d+\s`)
		sql = reg.ReplaceAllString(sql, "")

		// get columns info
		var ColInfo []Col
		if err = db.Select(&ColInfo, "SHOW FULL COLUMNS FROM "+tbName); err != nil {
			fmt.Println("Cannot get table [ " + tbName + " ] columns : " + err.Error())
			return
		}
		var tags []string
		if Tag != "" {
			tags = strings.Split(strings.ReplaceAll(Tag, " ", ""), ",")
		}

		var colList []string
		// beautiful code
		var maxTbLen int
		var maxTypeLen int
		for _, val := range ColInfo {
			colList = append(colList, val.Field)
			tbLen := len(fieldConvert(val.Field))
			if tbLen > maxTbLen {
				maxTbLen = tbLen
			}
			typeLen := len(colMatchList(val.Type, val.Null))
			if typeLen > maxTypeLen {
				maxTypeLen = typeLen
			}
		}

		// combine sql
		maxLen := 0
		type Arr struct {
			Key     string
			Comment string
		}
		var colsArr []Arr
		var pk string
		for _, val := range ColInfo {
			var ss string
			if len(tags) > 0 {
				for _, vv := range tags {

					ss += " " + tagInfo(val, vv)
				}
				ss = " `" + strings.Trim(ss, " ") + "`"
			}
			commentStr := ""
			if len(strings.TrimSpace(val.Comment)) > 0 {
				commentStr = " // " + val.Comment + "\n"
			}
			tmp := "    " + spaceFill(fieldConvert(val.Field), maxTbLen) + " " + spaceFill(colMatchList(val.Type, val.Null), maxTypeLen) + ss
			if maxLen < len(tmp) {
				maxLen = len(tmp)
			}
			colsArr = append(colsArr, Arr{tmp, commentStr})
			// only can recognize one primary key
			if pk == "" && val.Key == "PRI" {
				pk = val.Field
			}
		}
		colStr := ""
		for _, vv := range colsArr {
			if vv.Comment != "" {
				colStr += spaceFill(vv.Key, maxLen) + vv.Comment
			} else {
				colStr += vv.Key + "\n"
			}
		}
		mkdir(Dir)
		content := tpl(PackName, tbName, pk, sql, strings.Trim(colStr, "\n"), colList)
		fname := strings.TrimRight(Dir, "/") + "/" + PackName + "_" + tbName + ".go"
		if err := ioutil.WriteFile(fname, []byte(content), 0777); err != nil {
			fmt.Println(err.Error())
		}
	}
	if base == true {
		createBaseInterface(strings.TrimRight(Dir, "/")+"/"+PackName+"_base_interface.go", PackName)
	}
	scpt := ""
	fmt.Print("Create generate script? Y/y to generate, other omit : ")
	fmt.Scanln(&scpt)
	if scpt == "y" || scpt == "Y" {
		createScript(DBAddr, DBName, Dir, Tag, Adapter, TableFn, SqlNull)
	}
}

func checkArgs() bool {
	dir := flag.String("dir", "", "directory path")
	tag := flag.String("tag", "", "tags = xorm,json,db")
	adapter := flag.String("adapter", "", "for db adapter")
	sqlNull := flag.String("sqlnull", "", "for db sqlnull")
	dbAuth := flag.String("db_auth", "", "for db auth")
	dbAddr := flag.String("db_addr", "", "for db addr")
	dbName := flag.String("db_name", "", "for db name")
	fn := flag.String("fn", "", "for generate get table fun")
	tbList := flag.String("t", "", "for generate get table fun")

	flag.Parse()
	DBAddr = *dbAddr
	if *dbAddr == "" {
		fmt.Print("Input db address  '127.0.0.1:3306' or use ' -db_addr=127.0.0.1:3306 ' : ")
	DBADDR:
		fmt.Scanln(&DBAddr)
		if DBAddr == "" {
			fmt.Print("please input db address : ")
			goto DBADDR
		}
	}
	DBAuth = *dbAuth
	if *dbAuth == "" {
		fmt.Print("Input db auth 'root:123456' or use '-db_auth=root:123456' : ")
	DBAUTH:
		fmt.Scanln(&DBAuth)
		if DBAuth == "" {
			fmt.Print("please input db auth : ")
			goto DBAUTH
		}
	}
	// db name
	DBName = *dbName
	if *dbName == "" {
		fmt.Print("Input db name or use ' -db_name=your_db_name ' : ")
	DBNAME:
		fmt.Scanln(&DBName)
		if DBName == "" {
			fmt.Print("please input db name : ")
			goto DBNAME
		}
	}

	if *dir == "" {
		fmt.Print("Input directory path, default './model' or use ' -dir=./model ' : ")
		fmt.Scanln(dir)
		if *dir != "" {
			Dir = *dir
		} else {
			fmt.Println("using -dir=./model")
		}
	} else {
		Dir = *dir
	}

	// tag
	if *tag == "" {
		fmt.Print("Input any tags, use ',' to split it (empty will not create tag) or use -tags=xorm,yaml,json,db : ")
		fmt.Scanln(tag)
		if *tag != "" {
			Tag = *tag
		} else {
			fmt.Println("empty tag")
		}
	} else {
		Tag = *tag
	}

	if *adapter == "" {
		fmt.Print("Input adapter, 'postgres' or 'mysql', other default 'mysql', command : [ -adapter=postgres]  : ")
		fmt.Scanln(adapter)
		if *adapter != "" && (*adapter == "mysql" || *adapter == "postgres") {
			Adapter = *adapter
		} else {
			fmt.Println("using -adapter=mysql")
		}
	} else {
		Adapter = *adapter
	}

	if *sqlNull == "" {
		fmt.Print("Use sql.Null? Y/y to use, leave empty default no or use -sqlnull=y :")
		fmt.Scanln(sqlNull)
		if *sqlNull != "" && (*sqlNull == "Y" || *sqlNull == "y") {
			SqlNull = true
		} else {
			fmt.Println("not using sql.Null")
		}
	} else {
		if *sqlNull == "Y" || *sqlNull == "y" {
			SqlNull = true
		}
	}

	if *tbList == "" {
		fmt.Print("Input tables, others all tables, command [ '-t=table1,table2' ] : ")
		fmt.Scanln(tbList)
		if *tbList != "" {
			tbs := strings.Split(strings.ReplaceAll(strings.Trim(*tbList, ","), " ", ""), ",")
			if len(tbs) > 0 {
				TabList = tbs
			}
		}
	} else {
		tbs := strings.Split(strings.Trim(*tbList, ","), ",")
		TabList = tbs
	}

	if *fn == "" {
		fmt.Print("Generate get table name function? Y/y generate, other omit: ")
		fmt.Scanln(fn)
		if *fn == "y" || *fn == "Y" {
			TableFn = true
		}
	} else if *fn == "y" || *fn == "Y" {
		TableFn = true
	}

	// set package name
	if strings.Count(Dir, "/") > 0 {
		tmp := strings.Split(Dir, "/")
		PackName = tmp[len(tmp)-1]
	} else if strings.Count(Dir, "\\") > 0 {
		tmp := strings.Split(Dir, "\\")
		PackName = tmp[len(tmp)-1]
	} else {
		PackName = Dir
	}
	if *tbList == "" {
		return true
	}
	return false
}

func spaceFill(s string, i int) string {
	i = i - len(s)
	if i > 0 {
		return s + strings.Repeat(" ", i)
	}
	return s
}

func mkdir(dir string) {
	f, err := os.Stat(dir)
	if err == nil {
		return
	}
	if f != nil && !f.IsDir() {
		panic("exist same path file : " + dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			panic("Cannot mkdir for path: " + dir + " , error: " + err.Error())
		}
		return
	}
	panic("Cannot path error : " + dir + " , " + err.Error())
}

func fieldConvert(field string) string {
	// convert string to camel
	f := strcase.ToCamel(field)
	l := len(f)
	// convert id to ID
	if l == 2 {
		return strings.ToUpper(f)
	}
	if l == 3 && strings.ToLower(f[1:]) == "id" {
		return strings.ToUpper(f)
	}
	fieldLen := len(field)
	if l > 2 && fieldLen > 3 && field[fieldLen-3:] == "_id" {
		return f[:l-2] + "ID"
	}
	if fieldLen > 4 && field[fieldLen-4] == '_' && field[fieldLen-2:] == "id" {
		return f[:l-3] + strings.ToUpper(f[l-3:])
	}
	return f
}

func tpl(packName, tableName, pk, sql, colContent string, cols []string) string {
	tpl := `package {packName}
{import}
/*
{createTableSql}
*/

type {tableContent} struct {
{colContent}
}
{tableStr}`
	importTimeStr := `
import (
    "time"
)`
	importSqlStr := `
import (
    "database/sql"
)
`
	importTimeAndSqlStr := `
import (
    "database/sql"
    "time"
)
`
	tableStr := `
func (dst *{tableContent}) TableName() string {
    return "{tableName}"
}

func (dst *{tableContent}) PK() string {
    return "{pk}"
}

func (dst *{tableContent}) Cols() []string {
    return []string{{colsName}}
}
`
	tpl = strings.ReplaceAll(tpl, "{packName}", packName)
	tpl = strings.ReplaceAll(tpl, "{createTableSql}", sql)
	tpl = strings.ReplaceAll(tpl, "{colContent}", colContent)
	// import string
	if strings.Count(colContent, "time.Time") > 0 && strings.Count(colContent, "sql.Null") > 0 {
		tpl = strings.ReplaceAll(tpl, "{import}", importTimeAndSqlStr)
	} else if strings.Count(colContent, "time.Time") > 0 {
		tpl = strings.ReplaceAll(tpl, "{import}", importTimeStr)
	} else if strings.Count(colContent, "sql.Null") > 0 {
		tpl = strings.ReplaceAll(tpl, "{import}", importSqlStr)
	} else {
		tpl = strings.ReplaceAll(tpl, "{import}", "")
	}
	if TableFn {
		tpl = strings.ReplaceAll(tpl, "{tableStr}", tableStr)
		tpl = strings.ReplaceAll(tpl, "{tableName}", tableName)
	} else {
		tpl = strings.ReplaceAll(tpl, "{tableStr}", "")
	}
	tpl = strings.ReplaceAll(tpl, "{tableContent}", strcase.ToCamel(tableName))
	tpl = strings.ReplaceAll(tpl, "{pk}", pk)
	tpl = strings.ReplaceAll(tpl, "{colsName}", "\""+strings.Join(cols, "\", \"")+"\"")
	return tpl
}

// TypeMysqlDicMp Accurate matching type
var TypeMysqlDicMp = map[string]string{
	"bit(1)":             "[]uint8",
	"tinyint unsigned":   "uint8",
	"tinyint":            "int8",
	"smallint unsigned":  "uint16",
	"smallint":           "int16",
	"int unsigned":       "uint",
	"int":                "int",
	"bigint unsigned":    "uint64",
	"bigint":             "int64",
	"mediumint unsigned": "uint",
	"mediumint":          "int",
	"integer":            "int64",

	"float unsigned":  "float32",
	"float":           "float32",
	"double unsigned": "float64",
	"double":          "float64",

	"enum":       "string",
	"json":       "string",
	"varchar":    "string",
	"char":       "string",
	"date":       "string",
	"longtext":   "string",
	"tinytext":   "string",
	"mediumtext": "string",
	"text":       "string",

	"time":      "time.Time",
	"timestamp": "time.Time",
	"datetime":  "time.Time",

	"tinyblob":   "[]byte",
	"blob":       "[]byte",
	"mediumblob": "[]byte",
	"longblob":   "[]byte",
}

// TypeSqlNullList cause int,float,bool has a default value in golang, so they not in consider
var TypeSqlNullList = []struct {
	Key   string
	Value string
}{
	{`^(char)[(]\d+[)]`, `sql.NullString`},
	{`^(varchar)[(]\d+[)]`, `sql.NullString`},
	{`^(json)`, `sql.NullString`},
	{`^(text)`, `sql.NullString`},
	{`^(date)`, `sql.NullString`},
	{`^(longtext)`, `sql.NullString`},
	{`^(tinytext)`, `sql.NullString`},
	{`^(mediumtext)`, `sql.NullString`},

	{`^(time)`, `sql.NullTime`},
	{`^(timestamp)`, `sql.NullTime`},
	{`^(datetime)`, `sql.NullTime`},
}

var TypeMysqlMatchList = []struct {
	Key   string
	Value string
}{
	{`^(bit)[(]\d+[)]`, "[]uint8"},
	{`^(tinyint)[(]\d+[)] unsigned`, "uint8"},
	{`^(tinyint)[(]\d+[)]`, "int8"},
	{`^(smallint)[(]\d+[)] unsigned`, "uint16"},
	{`^(smallint)[(]\d+[)]`, "int16"},
	{`^(mediumint)[(]\d+[)]`, "string"},
	{`^(int)[(]\d+[)] unsigned`, "uint"},
	{`^(int)[(]\d+[)]`, "int"},
	{`^(bigint)[(]\d+[)] unsigned`, "uint64"},
	{`^(bigint)[(]\d+[)]`, "int64"},
	{`^(integer)[(]\d+[)]`, "int"},

	{`^(real)`, "float32"},
	{`^(float)[(]\d+,\d+[)] unsigned`, "float32"},
	{`^(float)[(]\d+,\d+[)]`, "float32"},
	{`^(double)[(]\d+,\d+[)] unsigned`, "float64"},
	{`^(double)[(]\d+,\d+[)]`, "float64"},
	{`^(decimal)[(]\d+,\d+[)]`, "float64"},
	{`^(numeric)[(]\d+,\d+[)]`, "float64"},

	{`^(enum)[(](.)+[)]`, "string"},
	{`^(char)[(]\d+[)]`, "string"},
	{`^(varchar)[(]\d+[)]`, "string"},
	{`^(text)[(]\d+[)]`, "string"},

	{`^(datetime)[(]\d+[)]`, "time.Time"},
	{`^(timestamp)[(]\d+[)]`, "time.Time"},

	{`^(blob)[(]\d+[)]`, "[]byte"},
	{`^(binary)[(]\d+[)]`, "[]byte"},
	{`^(varbinary)[(]\d+[)]`, "[]byte"},
	{`^(geometry)[(]\d+[)]`, "[]byte"},
}

func colMatchList(name string, isNull []byte) string {
	if !SqlNull && string(isNull) == "YES" {
		for _, l := range TypeSqlNullList {
			if ok, _ := regexp.MatchString(l.Key, name); ok {
				return l.Value
			}
		}
	}

	// Precise matching first
	if v, ok := TypeMysqlDicMp[name]; ok {
		return v
	}
	// Fuzzy Regular Matching
	for _, l := range TypeMysqlMatchList {
		if ok, _ := regexp.MatchString(l.Key, name); ok {
			return l.Value
		}
	}
	panic("type (" + name + ") not match in any way.")
}

func tagInfo(s Col, tag string) string {
	tagStr := ""
	tag = strings.ToLower(tag)
	if tag == "xorm" {

		tagStr = s.Field + " " + s.Type
		if s.Key != "" {
			if s.Key == "PRI" {
				tagStr = tagStr + " pk"
			}
			if s.Key == "UNI" {
				tagStr = " unique"
			}
		}
		if s.Extra == "auto_increment" {
			tagStr = tagStr + " autoincr"
		}
		if string(s.Null) == "NO" {
			tagStr = tagStr + " not null"
		}
		if string(s.Default) != "" {
			tagStr = tagStr + " default(" + string(s.Default) + ")"
		}
	} else if tag == "gorm" {
		if s.Key != "" {
			if s.Key == "PRI" {
				tagStr = "primary_key;"
			} else if s.Key == "UNI" {
				tagStr = "unique;"
			}
		}
		tagStr = tagStr + "column:" + s.Field + ";type:" + s.Type
		if string(s.Null) == "NO" {
			tagStr = tagStr + ";not null"
		}
		if string(s.Default) != "" {
			tagStr = tagStr + ";default:" + string(s.Default)
		}
	} else {
		tagStr = s.Field
	}

	return tag + ":\"" + tagStr + "\""
}

func createScript(addr, dbname, dir, tag, adpt string, tabFn, sqlNull bool) {
	var script, suffix string
	if "windows" == runtime.GOOS {
		suffix = ".bat"
		script = `dbtag -db_addr={db_addr} -db_name={dbname} -adapter={adpt} -dir={dir} {tags} -fn={fn} -sqlnull={sqlnull}
@pause`
	} else {
		suffix = ".sh"
		script = `#!/bin/bash
dbtag -db_addr={db_addr} -db_name={dbname} -adapter={adpt} -dir={dir} {tags} -fn={fn} -sqlnull={sqlnull}`
	}

	script = strings.ReplaceAll(script, "{db_addr}", addr)
	script = strings.ReplaceAll(script, "{dbname}", dbname)
	script = strings.ReplaceAll(script, "{adpt}", adpt)
	script = strings.ReplaceAll(script, "{dir}", dir)
	if tag != "" {
		tag = "-tag=" + tag
	}
	script = strings.ReplaceAll(script, "{tags}", tag)
	fn := "n"
	if tabFn {
		fn = "y"
	}
	script = strings.ReplaceAll(script, "{fn}", fn)
	nullStr := "n"
	if sqlNull {
		nullStr = "y"
	}
	script = strings.ReplaceAll(script, "{sqlnull}", nullStr)
	script = strings.ReplaceAll(script, "\\", "/")

	if err := ioutil.WriteFile("cmd_db"+suffix, []byte(script), 0777); err != nil {
		panic(err.Error())
	}
}

func createBaseInterface(fname, packageName string) {
	base := `package {packName}

type Model interface {
	TableName() string
	PK() string
	Cols() []string
}
`
	base = strings.ReplaceAll(base, "{packName}", packageName)
	if err := ioutil.WriteFile(fname, []byte(base), 0777); err != nil {
		fmt.Println(err.Error())
	}
}
