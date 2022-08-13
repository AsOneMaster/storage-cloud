package meta

type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
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
func GetFileMeta(filesha1 string) FileMeta {
	return fileMetas[filesha1]
}
