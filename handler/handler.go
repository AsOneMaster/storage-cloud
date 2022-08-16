package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"storage-cloud/meta"
	"storage-cloud/util"
	"strconv"
	"time"
)

const (
	UploadHtml          = "./static/view/index.html"
	TmpLocation         = "./static/tmp/"
	FileTypeHeader      = "Content-Type"
	FileTypeHeaderValue = "application/octect-stream"
	FileDisHeader       = "Content-Disposition"
	FileDisHeaderValue  = "attachment;filename=\""
)

//check err
func checkErr(err error, w http.ResponseWriter, errStr string) {
	if err != nil {
		io.WriteString(w, errStr)
		return
	}
}

// UploadHandler 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传页面
		data, err := ioutil.ReadFile(UploadHtml)
		checkErr(err, w, "filed to find html")
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流及存储本地目录
		file, head, err := r.FormFile("file")
		checkErr(err, w, "failed to get data")

		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName:   head.Filename,
			Location:   TmpLocation + head.Filename,
			UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		}
		//创建本地文件 接受文件流
		newFile, err := os.Create(fileMeta.Location)
		checkErr(err, w, "failed to create file")

		fileMeta.FileSize, err = io.Copy(newFile, file)
		checkErr(err, w, "failed to save data into file")
		//从头开始读取，不偏移
		newFile.Seek(0, io.SeekStart)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//新增文件到数据库
		ok := meta.UpLoadFileMetaDB(fileMeta)
		fmt.Println(fileMeta.FileSha1)
		if ok {
			http.Redirect(w, r, "/file/upload/success", http.StatusFound)
		} else {
			w.Write([]byte("上传失败"))
		}

	}

}

// UploadSuccessHandler 上传完成
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload Success!")
}

// GetFileMetaHandler 获取文件元信息 curl: http://localhost:8080/file/meta?fileHash=
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	//ParseForm 填充 r.Form 和 r.PostForm。
	//对于所有请求，ParseForm 解析来自 URL 的原始查询并更新 r.Form。
	//对于 POST、PUT 和 PATCH 请求，它还会读取请求正文，将其解析为表单并将结果放入 r.PostForm 和 r.Form 中。 请求正文参数优先于 r.Form 中的 URL 查询字符串值。
	r.ParseForm()

	fileHash := r.Form["fileHash"][0]
	fileMeta, err := meta.GetFileMeta(fileHash)
	checkErr(err, w, "云文件数据库查询失败")
	data, err := json.Marshal(fileMeta)
	checkErr(err, w, "转换json失败")
	w.Write(data)
}

// FileQueryHandler 批量获取文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitStr := r.Form.Get("limit")
	limitInt, err := strconv.Atoi(limitStr)
	checkErr(err, w, "limitStr转换成int失败")
	fileMetas := meta.GetLastFileMetas(limitInt)
	data, err := json.Marshal(fileMetas)
	checkErr(err, w, "批量文件元信息json转换失败")
	w.Write(data)
}

// DownloadHandler 文件下载  curl: http://localhost:8080/file/download?fileHash=
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("fileHash")
	//获取fileMeta对象信息
	fm, err := meta.GetFileMeta(fileSha1)
	checkErr(err, w, "云文件数据库查询失败")
	openFile, err := os.Open(fm.Location)
	checkErr(err, w, "云文件地址获取失败")
	defer openFile.Close()
	data, err := ioutil.ReadAll(openFile)
	checkErr(err, w, "云文件读取失败")

	//识别http浏览器响应头 让浏览器识别出是文件下载
	w.Header().Set(FileTypeHeader, FileTypeHeaderValue)
	w.Header().Set(FileDisHeader, FileDisHeaderValue+fm.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler 更新元信息（重命名） curl: http://localhost:8080/file/update?op= &&fileHash= &&fileName=   [注意使用：POST请求测试]
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//文件操作暗号
	opType := r.Form.Get("op")

	//获取文件sha1值
	fileSha1 := r.Form.Get("fileHash")

	//获取更新文件位置
	newFileLocation := r.Form.Get("fileAddr")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//根据sha1获取文件元信息
	curFileMeta, err := meta.GetFileMeta(fileSha1)
	checkErr(err, w, "云文件数据库查询失败")
	//重命名文件名
	curFileMeta.Location = newFileLocation
	//curFileMeta.Location = TmpLocation + newFileName
	//更新文件元信息
	meta.UpdateFileMetaLocationDB(curFileMeta)
	//转换为json值
	data, err := json.Marshal(curFileMeta)

	checkErr(err, w, "json转换失败")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler 文件删除 curl: http://localhost:8080/file/update?fileSha1=
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("fileSha1")

	fileMeta, err := meta.GetFileMeta(fileSha1)
	checkErr(err, w, "云文件数据库查询失败")
	//fmt.Println(meta.GetLastFileMetas(2))
	err = os.Remove(fileMeta.Location)
	checkErr(err, w, "删除不成功")
	meta.RemoveFileMeta(fileSha1)
	//fmt.Println(meta.GetLastFileMetas(2))
	w.WriteHeader(http.StatusOK)
}
