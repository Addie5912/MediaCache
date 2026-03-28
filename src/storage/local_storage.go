package storage

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// VideoCacheDirPermission 缓存目录权限
	VideoCacheDirPermission = 0755
)

// localStorage 本地存储实现
type localStorage struct {
	name     string // 存储器名称
	basePath string // 基础路径，缓存根目录
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(name string) *localStorage {
	return &localStorage{
		name:     name,
		basePath: conf.Instance().MediaCache,
	}
}

// Cache 为指定文件创建缓存文件，返回可写入的文件信息
func (storage *localStorage) Cache(filePath string) (*FileInfo, error) {
	if !conf.IsCacheAvailable() {
		return nil, fmt.Errorf("cache is not available")
	}

	absolutePath := filepath.Join(storage.basePath, filePath)
	// 确保存储目录存在
	if err := os.MkdirAll(filepath.Dir(absolutePath), VideoCacheDirPermission); err != nil {
		return nil, fmt.Errorf("[localStorage] failed to create directory[%s], err: %v", filepath.Dir(absolutePath), err)
	}

	// 以写入模式打开文件
	file, err := os.Create(absolutePath)
	if err != nil {
		storage.clean(filePath)
		return nil, fmt.Errorf("[localStorage] failed to create file[%s], err: %v", absolutePath, err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		storage.clean(filePath)
		logger.Errorf("[localStorage] Seek file[%s] failed, error: %v", absolutePath, err)
		return nil, err
	}

	return &FileInfo{
		Name:             filepath.Base(absolutePath),
		Path:             absolutePath,
		ModifiedTime:     time.Now(),
		ExtraWriteTarget: file,
		HasCached:        false,
		Finalizer: func() {
			file.Close()
		},
	}, nil
}

// Get 从本地缓存中读取文件内容，并获取文件元数据
func (storage *localStorage) Get(videoPath string) (io.ReadCloser, *FileInfo, error) {
	absolutePath := filepath.Join(storage.basePath, videoPath)
	logger.Infof("[localStorage] will get file[%s] from local cache", absolutePath)

	file, err := os.Open(absolutePath)
	if err != nil {
		logger.Errorf("[localStorage] get file[%s] failed, error: %v", absolutePath, err)
		return nil, nil, err
	}

	localFileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("[localStorage] stat file[%s] failed, error: %v", absolutePath, err)
		return nil, nil, err
	}

	hash := storage.generateHash(file, absolutePath)
	if hash == "" {
		logger.Errorf("[localStorage] generateHash[%s] failed, error: %v", absolutePath, err)
		return nil, nil, err
	}

	fileInfo := &FileInfo{
		Name:         filepath.Base(absolutePath),
		Path:         absolutePath,
		Size:         strconv.FormatInt(localFileInfo.Size(), 10),
		ModifiedTime: localFileInfo.ModTime(),
		Hash:         hash,
	}

	return file, fileInfo, nil
}

// Exist 检查指定文件是否存在于本地缓存中
func (storage *localStorage) Exist(videoPath string) bool {
	absolutePath := filepath.Join(storage.basePath, videoPath)

	_, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("[localStorage] file[%s] not exist in localStorage", absolutePath)
			return false
		}
		logger.Errorf("[localStorage] failed to check file[%s] existence: %v", absolutePath, err)
		return false
	}

	return true
}

// clean 删除指定缓存文件，用于清理操作和失败回滚
func (storage *localStorage) clean(videoPath string) {
	absolutePath := filepath.Join(storage.basePath, videoPath)
	if err := os.Remove(absolutePath); err != nil {
		logger.Errorf("[localStorage] Failed to delete file[%s]: %v", absolutePath, err)
	}
}

// generateHash 计算文件 MD5 哈希值（base64 编码），计算后将文件指针重置到开头
func (storage *localStorage) generateHash(file *os.File, filePath string) string {
	hashEncoder := md5.New()
	if _, err := io.Copy(hashEncoder, file); err != nil {
		logger.Errorf("[localStorage] hash file[%s] failed, error: %v", filePath, err)
		return ""
	}

	hash := base64.StdEncoding.EncodeToString(hashEncoder.Sum(nil)[:])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logger.Errorf("[localStorage] Seek file[%s] failed, error: %v", filePath, err)
		return ""
	}

	return hash
}
