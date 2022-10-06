package cos

import (
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"storage-cloud/config"
)

var cosCli *cos.Client

// Client : 创建oss client对象
func Client() *cos.Client {
	if cosCli != nil {
		return cosCli
	}
	u, _ := url.Parse(config.BucketURL)
	// 用于Get Service 查询，默认全地域 service.cos.myqcloud.com
	su, _ := url.Parse(config.ServiceURL)
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	// 1.永久密钥
	cosCli = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SECRETID,  // 替换为用户的 SecretId，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
			SecretKey: config.SecretKey, // 替换为用户的 SecretKey，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
		},
	})
	return cosCli
}
