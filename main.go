package main

import (
	"fmt"
	"net/http"
	"storage-cloud/handler"
)

func main() {
	// 静态资源处理
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static"))))

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

	// 秒传接口
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(
		handler.TryFastUploadHandler))
	//
	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(
		handler.DownloadURLHandler))

	// 用户相关接口
	// http.HandleFunc("/", handler.SignInHandler)
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	// 监听端口
	fmt.Println("上传服务正在启动, 监听端口:8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}
