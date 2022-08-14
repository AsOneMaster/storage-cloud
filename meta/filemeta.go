package meta

import "sort"

type FileMeta struct {
	FileSha1   string
	FileName   string
	FileSize   int64
	Location   string
	UploadTime string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta 新增/更新元信息
func UpdateFileMeta(meta FileMeta) {
	fileMetas[meta.FileSha1] = meta
}

// GetFileMeta 通过sha1值获取文件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetLastFileMetas 批量获取文件的元信息对象
func GetLastFileMetas(count int) []FileMeta {
	fileMetaArray := make([]FileMeta, len(fileMetas))
	for _, fileMeta := range fileMetas {
		fileMetaArray = append(fileMetaArray, fileMeta)
	}
	//使用sort排序 会调用类型的 less len swap方法 ByUploadTime重写了该三种方法 适用于该file类型比较time
	sort.Sort(ByUploadTime(fileMetaArray))
	return fileMetaArray[0:count]
}
