package remote

import (
	"mediaCacheService/storage"
	"io"
)

// Remote 远程服务接口（接口签名保持不变）
type Remote interface {
	GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
	PostValidateIMEI(IMEI string, CheckType string) (bool, error)
	GetGIDSAddress() (string, error)
}
