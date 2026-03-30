package tasks

import (
	goctx "context"
	"math"
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"mediaCacheService/util/sys"
	"strconv"
)

// scanningTask 扫描任务：统计缓存目录磁盘使用情况，更新缓存可用状态
func scanningTask(ctx goctx.Context) error {
	select {
	case <-ctx.Done():
		logger.Infof("[ScanningTask] task is over")
		return nil
	default:
	}

	cf := conf.Instance()
	cachePath := cf.MediaCache

	sysI := sys.NewFunc()
	usedMb, err := sysI.SysDirSize(cachePath)
	if err != nil {
		logger.Errorf("[ScanningTask] SysDirSize failed: %v", err)
		return err
	}

	// 解析阈值（MB），计算 85% 可用阈值
	cleanOriThreshold, err := strconv.Atoi(cf.DataAging.ClearingTaskThreshold)
	if err != nil {
		logger.Errorf("[ScanningTask] parse ClearingTaskThreshold failed: %v", err)
		return err
	}
	threshold := int(math.Floor(float64(cleanOriThreshold) * 0.85))

	available := usedMb <= threshold
	conf.SetCacheAvailable(available)
	logger.Infof("[ScanningTask] success scanning dir size, current cache available stat is: %v (usedMb=%d, threshold=%d)",
		available, usedMb, threshold)
	return nil
}
