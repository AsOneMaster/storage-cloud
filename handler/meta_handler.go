package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"storage-cloud/config"
	dblayer "storage-cloud/db"
	"storage-cloud/meta"
	"storage-cloud/store/cos"
	"storage-cloud/util"
	"strconv"
	"time"
)

const (
	UploadHtml          = "./static/view/index.html"
	TmpLocation         = "./static/tmp/"
	COSLocation         = "cos/"
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
		//file, head, err := r.FormFile("file")
		file, head, err := r.FormFile("file")
		checkErr(err, w, "failed to get data")

		defer file.Close()

		//创建本地文件 接受文件流
		//newFile, err := os.Create(fileMeta.Location)
		fileMeta := meta.FileMeta{
			FileName:   head.Filename,
			Location:   TmpLocation + head.Filename,
			UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err:%s\n", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err:%s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		// 游标重新回到文件头部
		newFile.Seek(0, 0)

		//腾讯云COS上传
		c := cos.Client()
		cosPath := COSLocation + fileMeta.FileSha1
		// 1. 获取预签名URL
		presignedURL, err := c.Object.GetPresignedURL(context.Background(), http.MethodPut, cosPath, config.SECRETID, config.SecretKey, time.Hour, nil)
		if err != nil {
			panic(err)
		}
		// 2. 通过预签名方式上传对象
		req, err := http.NewRequest(http.MethodPut, presignedURL.String(), newFile)
		if err != nil {
			panic(err)
		}
		// 用户可自行设置请求头部
		contentType := head.Header.Get("Content-Type")

		req.Header.Set("Content-Type", contentType)
		//使浏览器不自动预览
		req.Header.Set("Content-Disposition", "attachment")
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		//_, err = c.Object.Put(context.Background(), cosPath, newFile, nil)
		//if err != nil {
		//	panic(err)
		//}
		checkErr(err, w, "failed to create file")
		fileMeta.Location = cosPath
		//新增文件到数据库
		_ = meta.UpLoadFileMetaDB(fileMeta)
		// 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		ok := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		//fmt.Println(fileMeta.FileSha1)
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
	//r.ParseForm()
	//limitStr := r.Form.Get("limit")
	//limitInt, err := strconv.Atoi(limitStr)
	//checkErr(err, w, "limitStr转换成int失败")
	//fileMetas := meta.GetLastFileMetas(limitInt)
	//data, err := json.Marshal(fileMetas)
	//checkErr(err, w, "批量文件元信息json转换失败")
	//w.Write(data)

	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Println("data--------------------", string(data))
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
	meta.UpdateFileMetaLocationDB(*curFileMeta)
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

// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMeta(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

// DownloadURLHandler : 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)
	//// 1.获取对象访问 URL
	cosPath := row.FileAddr.String
	c := cos.Client()
	// 获取预签名URL
	presignedURL, err := c.Object.GetPresignedURL(context.Background(), http.MethodGet, cosPath, config.SECRETID, config.SecretKey, time.Hour, nil)
	if err != nil {
		panic(err)
	}
	//通过预签名URL下载对象
	w.Write([]byte(presignedURL.String()))

	//-------------------------------
	// 1. 从响应体中获取对象
	//resp, err := c.Object.Get(context.Background(), cosPath, nil)
	//if err != nil {
	//	panic(err)
	//}
	//data, err := ioutil.ReadAll(resp.Body)
	//checkErr(err, w, "云文件读取失败")
	//
	////识别http浏览器响应头 让浏览器识别出是文件下载
	//w.Header().Set(FileTypeHeader, FileTypeHeaderValue)
	//w.Header().Set(FileDisHeader, FileDisHeaderValue+row.FileName.String+"\"")
	//w.Write(data)
	//-------------------------------------

	//// TODO: 判断文件存在OSS，还是Ceph，还是在本地
	//if strings.HasPrefix(row.FileAddr.String, "/tmp") {
	//	username := r.Form.Get("username")
	//	token := r.Form.Get("token")
	//	tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
	//		r.Host, filehash, username, token)
	//	w.Write([]byte(tmpUrl))
	//} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
	//	// TODO: ceph下载url
	//} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
	//	// oss下载url
	//	signedURL := oss.DownloadURL(row.FileAddr.String)
	//	w.Write([]byte(signedURL))
	//}
}
