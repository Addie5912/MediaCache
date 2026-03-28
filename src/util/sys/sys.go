package sys

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Interface 系统操作接口
type Interface interface {
	SysDirSize(dirPath string) (int64, error)
	DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
}

// sysImpl 系统操作实现
type sysImpl struct{}

// NewFunc 创建系统操作实例
func NewFunc() Interface {
	return &sysImpl{}
}

// SysDirSize 获取目录大小（MB），使用 du -sm 命令
func (s *sysImpl) SysDirSize(dirPath string) (int64, error) {
	cmd := exec.Command("du", "-sm", dirPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("du -sm failed: %v", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return 0, fmt.Errorf("du -sm output is empty")
	}
	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse du output failed: %v", err)
	}
	return size, nil
}

// DeleteInactiveFile 删除访问时间超过阈值天数的文件，支持超时控制
func (s *sysImpl) DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c",
		fmt.Sprintf("find %s -type f -atime +%d -exec rm -f {} \\;", dirPath, threshold))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete inactive files: %v, output: %s", err, output)
	}
	return nil
}

// GetEnv 从环境变量获取值，不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// GetAppID 获取应用ID
func GetAppID() string {
	return GetEnv("APPID", "")
}

// GetPodName 获取Pod名称
func GetPodName() string {
	return GetEnv("PODNAME", "")
}

// GetNamespace 获取命名空间
func GetNamespace() string {
	return GetEnv("NAMESPACE", "")
}

// GetNodeName 获取节点名称
func GetNodeName() string {
	return GetEnv("NODENAME", "")
}

// GetServiceName 获取服务名称
func GetServiceName() string {
	return GetEnv("SERVICENAME", "MediaCacheService")
}

// IsHTTPEnabled 是否启用HTTP服务
func IsHTTPEnabled() bool {
	return GetEnv("ENABLE_HTTP", "true") == "true"
}
