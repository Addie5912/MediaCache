package storage

import (
	"io"
	"time"
)

// Storage 存储接口（接口签名保持不变）
type Storage interface {
	Cache(filePath string) (*FileInfo, error)
	Get(videoPath string) (io.ReadCloser, *FileInfo, error)
	Exist(filePath string) bool
}

// FileInfo 文件元数据模型（字段保持不变）
type FileInfo struct {
	Name             string
	Path             string
	Size             string
	ModifiedTime     time.Time
	Hash             string
	HasCached        bool
	ExtraWriteTarget io.WriteSeeker
	Finalizer        func()
}

// AddFinalizer 添加资源清理回调（保持一致性）
func (f *FileInfo) AddFinalizer(newFinalizer func()) {
	oldFinalizer := f.Finalizer
	f.Finalizer = func() {
		if oldFinalizer != nil {
			oldFinalizer()
		}
		newFinalizer()
	}
}
