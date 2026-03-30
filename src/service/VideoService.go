package service

import (
	"fmt"
	"io"
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"mediaCacheService/remote"
	"mediaCacheService/storage"
	"os"
	"path/filepath"
	"time"
)

// VideoService 视频服务接口
type VideoService interface {
	GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
	Download(path string) (io.ReadCloser, int64, error)
}

// VideoServiceImpl 视频服务实现结构体
type VideoServiceImpl struct {
	remote  remote.Remote   // 远程服务接口
	storage storage.Storage // 存储服务接口
	alarm   AlarmService    // 告警服务接口
}

// NewVideoService 创建视频服务实例（依赖注入）
func NewVideoService() *VideoServiceImpl {
	return &VideoServiceImpl{
		remote:  remote.NewRemote(),
		storage: storage.NewLocalStorage(storage.LocalStorage),
		alarm:   NewAlarmService(),
	}
}

// GetVideo 获取视频文件，优先从本地缓存获取，缓存不存在时从远程下载并写入缓存
func (v *VideoServiceImpl) GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error) {
	startTime := time.Now()
	videoPath = filepath.Clean(videoPath)

	// 检查本地缓存是否存在, 判断缓存可用是阔写出来的
	if conf.IsCacheAvailable() && v.storage.Exist(videoPath) {
		reader, fileInfo, err := v.storage.Get(videoPath)
		if err == nil {
			logger.Infof("[VideoService] cache hit for video: %s", videoPath, time.Since(startTime))
			// v.alarm.ClearAlarm(AlarmId300020, "get video from local cache success")
			return reader, fileInfo, nil
		}
		logger.Warnf("[VideoService] cache get failed, fallback to remote: %s, err: %v", videoPath, err)
	}

	// 缓存未命中，从远程MUEN服务下载
	remoteReader, remoteFileInfo, err := v.remote.GetVideo(videoPath)
	if err != nil {
		logger.Errorf("[VideoService] get video from remote failed, err: %v", err)
		v.alarm.SendAlarm(AlarmId300020, "Failed to get video content "+videoPath)
		return nil, nil, err
	}

	// 远程获取成功，清除告警
	v.alarm.ClearAlarm(AlarmId300020, "get video from MUEN success")

	// 将远程流写入本地缓存（同时返回给调用者）
	if conf.IsCacheAvailable() {
		cacheFileInfo, cacheErr := v.storage.Cache(videoPath)
		if cacheErr != nil {
			// 处理失败的分支
			logger.Warnf("[VideoService] create cache file failed: %v, skip caching", cacheErr)
			return remoteReader, remoteFileInfo, cacheErr
		}

		// 使用 TeeReader 同时写入缓存和返回给调用者
		//teeReader := io.TeeReader(remoteReader, cacheFileInfo.ExtraWriteTarget)
		//if remoteFileInfo != nil {
		//	remoteFileInfo.AddFinalizer(func() {
		//		remoteReader.Close()
		//		if cacheFileInfo.Finalizer != nil {
		//			cacheFileInfo.Finalizer()
		//		}
		//	})
		//}
		cacheFileInfo.Hash = remoteFileInfo.Hash
		cacheFileInfo.AddFinalizer(remoteFileInfo.Finalizer)

		// return io.NopCloser(teeReader), remoteFileInfo, nil
	}

	return remoteReader, remoteFileInfo, nil
}

// Download 直接从本地文件系统下载视频，不经过缓存机制
//func (v *VideoServiceImpl) Download(videoPath string) (io.ReadCloser, int64, error) {
//	if !v.storage.Exist(videoPath) {
//		return nil, 0, fmt.Errorf("[VideoService] file not found: %s", videoPath)
//	}
//
//	reader, fileInfo, err := v.storage.Get(videoPath)
//	if err != nil {
//		return nil, 0, fmt.Errorf("[VideoService] open file failed: %w", err)
//	}
//
//	var size int64
//	if fileInfo != nil && fileInfo.Size != "" {
//		fmt.Sscanf(fileInfo.Size, "%d", &size)
//	}
//	return reader, size, nil
//}

func (v *VideoServiceImpl) Download(videoPath string) (io.ReadCloser, int64, error) {
	fi, err := os.Stat("/" + videoPath)
	if os.IsNotExist(err) {
		return nil, 0, fmt.Errorf("videoPath %s does not exist", videoPath)
	}

	file, err := os.Open(videoPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file %s, err: %v", videoPath, err)
	}

	return file, fi.Size(), nil
}
