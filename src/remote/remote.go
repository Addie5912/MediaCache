package remote

import (
	"io"
	"mediaCacheService/storage"
)

type MicroServiceKey struct {
	AppId       string
	ServiceName string
	Version     string
}

var (
	// MUENMediaMouduleAddress muen video address example https://999.460.ylplo.ts.tmofamily.com:40050/api/%s
	//MUENMediaMouduleAddress = os.Getenv("MUEN_MEDIA_URL_PREFIX")
	// mock: 存量代码在获取muen地址的时候没有做获取不到的判断，改写的做了，所以要打桩
	MUENMediaMouduleAddress = "https://999.460.ylplo.ts.tmofamily.com:40050/api/%s"
)

// Remote 远程服务接口（接口签名保持不变）
type Remote interface {
	GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
	PostValidateIMEI(IMEI string, CheckType string) (bool, error)
	GetGIDSAddress() (string, error)
}
