package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/cloud?charset=utf8")
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
