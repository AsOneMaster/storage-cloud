package meta

import (
	"fmt"
	"sort"
	"storage-cloud/db"
)

type FileMeta struct {
	FileSha1   string
	FileName   string
	FileSize   int64
	Location   string
	UploadTime string
}

//UpdateFileMetaLocationDB 新增/更新元信息到数据库
func UpdateFileMetaLocationDB(meta FileMeta) bool {
	//fileMetas[meta.FileSha1] = meta
	return db.UpdateFileLocation(meta.FileSha1, meta.Location)
}

// UpLoadFileMetaDB 新增元信息到数据库
func UpLoadFileMetaDB(meta FileMeta) bool {
	return db.OpenFileUploadFinished(meta.FileSha1, meta.FileName, meta.FileSize, meta.Location)
}

// GetFileMeta 通过sha1值获取文件的元信息对象
func GetFileMeta(fileSha1 string) (*FileMeta, error) {
	tableFile, err := db.GetFileMeta(fileSha1)
	file := FileMeta{
		FileSha1:   tableFile.FileHash,
		FileName:   tableFile.FileName.String,
		FileSize:   tableFile.FileSize.Int64,
		Location:   tableFile.FileAddr.String,
		UploadTime: tableFile.UploadTime,
	}
	return &file, err
}

// GetLastFileMetas 批量获取文件的元信息对象
func GetLastFileMetas(count int) *[]FileMeta {
	fileMetaArray, err := db.GetFileMetaList(count)
	if err != nil {
		fmt.Println("批量获取数据库文件信息失败")
	}
	var fileMetas []FileMeta
	for _, fileMeta := range fileMetaArray {
		file := &FileMeta{
			FileSha1:   fileMeta.FileHash,
			FileName:   fileMeta.FileName.String,
			FileSize:   fileMeta.FileSize.Int64,
			Location:   fileMeta.FileAddr.String,
			UploadTime: fileMeta.UploadTime,
		}
		fileMetas = append(fileMetas, *file)
	}
	//使用sort排序 会调用类型的 less len swap方法 ByUploadTime重写了该三种方法 适用于该file类型比较time
	sort.Sort(ByUploadTime(fileMetas))
	return &fileMetas
}

// RemoveFileMeta 删除元信息
func RemoveFileMeta(fileSha1 string) {
	//如果线程同步 要考虑安全问题 加锁
	//delete(fileMetas, fileSha1)
}
