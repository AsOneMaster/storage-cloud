package main

import (
	"fmt"
	"net/http"
	"storage-cloud/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/success", handler.UploadSuccessHandler)
	//获取文件元信息 sha1
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	//获取批量文件元信息
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	//下载文件
	http.HandleFunc("/file/download", handler.DownloadHandler)
	//重命名文件
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	//删除文件
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}
