package tasks

import (
	"fmt"
	goctx "context"
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"strconv"

	"github.com/beego/beego/v2/task"
)

// InitCronTasks 初始化并注册所有定时任务
func InitCronTasks() {
	cf := conf.Instance()

	// 解析扫描任务周期（默认1小时）
	stp, err := strconv.Atoi(cf.DataAging.ScanningTaskPeriod)
	if err != nil {
		logger.Errorf("[Tasks] parse ScanningTaskPeriod failed: %v, use default 1h", err)
		stp = 1
	}

	// 解析清理任务周期（默认24小时）
	ctp, err := strconv.Atoi(cf.DataAging.ClearingTaskPeriod)
	if err != nil {
		logger.Errorf("[Tasks] parse ClearingTaskPeriod failed: %v, use default 24h", err)
		ctp = 24
	}

	// 注册扫描任务（每 stp 小时执行一次）
	task.AddTask("scanning_task", task.NewTask("scanning_task",
		fmt.Sprintf("0 0 */%d * * *", stp), scanningTask))

	// 注册清理任务（每 ctp 小时执行一次）
	task.AddTask("clearing_task", task.NewTask("clearing_task",
		fmt.Sprintf("0 0 */%d * * *", ctp), clearingTask))

	// 程序启动时立即执行一次清理任务
	if err := clearingTask(goctx.TODO()); err != nil {
		logger.Errorf("[Tasks] startup clearingTask failed: %v", err)
	}
}
