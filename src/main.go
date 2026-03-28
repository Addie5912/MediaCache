package main

import (
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"mediaCacheService/routers"
	"mediaCacheService/tasks"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. 加载配置
	conf.Init()

	// 2. 初始化日志
	logger.Init(conf.GlobalConfig.Logger.LogFile, conf.GlobalConfig.Logger.LogLevel)

	// 3. 初始化定时任务
	tasks.Init()

	// 4. 注册路由并启动HTTP服务
	routers.Init()

	// 5. 等待退出信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("MediaCacheService shutting down...")
}
