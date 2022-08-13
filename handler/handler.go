package handler

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"storage-cloud/meta"
	"storage-cloud/util"
	"time"
)

const (
	UploadHtml  = "./static/view/index.html"
	TmpLocation = "./static/tmp/"
)

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
			FileName: head.Filename,
			Location: TmpLocation + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		//创建本地文件 接受文件流
		newFile, err := os.Create(fileMeta.Location)
		checkErr(err, w, "failed to create file")

		fileMeta.FileSize, err = io.Copy(newFile, file)
		checkErr(err, w, "failed to save data into file")

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMeta(fileMeta)

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
	}

}

// UploadSuccessHandler 上传完成
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload Success!")
}
func checkErr(err error, w http.ResponseWriter, errStr string) {
	if err != nil {
		io.WriteString(w, errStr)
		return
	}
}
