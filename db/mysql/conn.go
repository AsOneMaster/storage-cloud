package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	cfg "storage-cloud/config"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", cfg.MySQLSource)
	//同时活跃的连接数
	db.SetMaxOpenConns(1000)
	//测试连接
	err := db.Ping()
	if err != nil {
		fmt.Println("Filed to connect")
		os.Exit(1)
	}
}

// DBConn 返回数据连接对象
func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
