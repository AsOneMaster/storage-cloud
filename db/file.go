package db

import (
	"database/sql"
	"fmt"
	"storage-cloud/db/mysql"
)

// OpenFileUploadFinished 文件上传完成，保存meta
func OpenFileUploadFinished(fileHash string, fileName string, fileSize int64, fileAddr string) bool {
	//通过per stmt-防止sql注入攻击
	sqlStr := "insert  into tbl_file (`file_sha1`,`file_name`,`file_size`,`file_addr`,`status`) values (?,?,?,?,1)"
	stmt, err := mysql.DBConn().Prepare(sqlStr)
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()
	ret, err := stmt.Exec(fileHash, fileName, fileSize, fileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	//RowsAffected() 返回插入行数量
	if rf, err := ret.RowsAffected(); nil == err {
		//插入失败  没有新纪录
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", fileHash)
		}
		return true
	}
	return false
}

// TableFile : 文件表结构体
type TableFile struct {
	FileHash   string
	FileName   sql.NullString
	FileSize   sql.NullInt64
	FileAddr   sql.NullString
	UploadTime string
}

// GetFileMeta : 从mysql获取文件元信息
func GetFileMeta(fileHash string) (*TableFile, error) {
	sqlStr := "select `file_sha1`,`file_addr`,`file_name`,`file_size`,`create_at` from tbl_file where `file_sha1`=? and `status`=1 limit 1"
	stmt, err := mysql.DBConn().Prepare(sqlStr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(fileHash).Scan(
		&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize, &tfile.UploadTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录， 返回参数及错误均为nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	//取结构体指针 是因为只生成一个内存地址，如果为值类型 函数返回时会将该值拷贝给新接受变量，新变量会从新分配内存空间
	return &tfile, nil
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) (tableFiles []*TableFile, err error) {
	sqlStr := "select `file_sha1`,`file_addr`,`file_name`,`file_size`,`create_at` from tbl_file where  `status`=1 limit ?"
	stmt, err := mysql.DBConn().Prepare(sqlStr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()
	//获取多行数据 查询使用给定参数执行准备好的查询语句，并将查询结果作为*row类型返回。
	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	//Columns返回列名。如果行关闭，列将返回错误。
	//cloumns, err := rows.Columns()
	//RawBytes是一个字节片，它保存对数据库本身拥有的内存的引用。在对RawBytes进行扫描后，切片仅在下次调用next、Scan或Close之前有效
	//values := make([]sql.RawBytes, len(cloumns))
	// 循环读取数据
	for rows.Next() {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize, &tfile.UploadTime)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
		}
		fmt.Printf("tfile:%#v\n", tfile)
		tableFiles = append(tableFiles, &tfile)
	}
	return
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(fileHash string, fileAddr string) bool {
	sqlStr := "update tbl_file set`file_addr`=? where `file_sha1`=?"
	stmt, err := mysql.DBConn().Prepare(sqlStr)
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileAddr, fileHash)
	fmt.Println(ret.RowsAffected())
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, fileHash:%s\n", fileHash)
		}
		return true
	}
	return false
}
