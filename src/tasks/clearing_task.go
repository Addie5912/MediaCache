package tasks

import (
	goctx "context"
	"math"
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"mediaCacheService/util/sys"
	"strconv"
	"time"
)

// clearingTask 清理任务：删除访问时间超过阈值的不活跃缓存文件
func clearingTask(ctx goctx.Context) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		logger.Infof("[ClearingTask] tasks is stop")
		return nil
	default:
	}

	cf := conf.Instance()
	cachePath := cf.MediaCache

	// 解析清理阈值，计算触发阈值（原阈值的75%）
	cleanOriThreshold, err := strconv.Atoi(cf.DataAging.ClearingTaskThreshold)
	if err != nil {
		logger.Errorf("[ClearingTask] parse ClearingTaskThreshold failed: %v", err)
		return err
	}
	cleanThreshold := int(math.Floor(float64(cleanOriThreshold) * 0.75))

	// 获取当前目录使用大小（MB）
	sysI := sys.NewFunc()
	usedMb, err := sysI.SysDirSize(cachePath)
	if err != nil {
		logger.Errorf("[ClearingTask] SysDirSize failed: %v", err)
		return err
	}

	// 未超过清理阈值则跳过
	if usedMb <= int64(cleanThreshold) {
		logger.Infof("[ClearingTask] usedMb %d <= cleanThreshold %d, skip cleaning task", usedMb, cleanThreshold)
		return nil
	}

	// 解析文件不活跃阈值（天数）
	threshold, err := strconv.Atoi(cf.DataAging.FileAccessInactiveThreshold)
	if err != nil {
		logger.Errorf("[ClearingTask] parse FileAccessInactiveThreshold failed: %v", err)
		return err
	}

	// 解析删除操作超时时间（秒）
	timeoutSec, err := strconv.Atoi(cf.DataAging.DeleteInactiveFileTimeout)
	if err != nil {
		logger.Errorf("[ClearingTask] parse DeleteInactiveFileTimeout failed: %v", err)
		return err
	}

	// 执行删除不活跃文件
	if err = sysI.DeleteInactiveFile(cachePath, threshold, time.Duration(timeoutSec)*time.Second); err != nil {
		logger.Errorf("[ClearingTask] failed to exec DeleteInactiveFile: %v", err)
		return err
	}
	logger.Infof("[ClearingTask] success delete inactive file, threshold is %d", threshold)

	// 清理完成后主动调用扫描任务更新缓存可用状态
	return scanningTask(ctx)
}
