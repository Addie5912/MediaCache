package conf

import (
	"os"
	"sync"
)

// Config 全局配置结构体
type Config struct {
	Logger      LoggerConfig    `flag:"log"`
	MediaCache  string          `flag:"cache" desc:"video local cache address"`
	HTTPTimeout int
	DataAging   DataAgingConfig `flag:"data"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	LogFile  string `flag:"file" desc:"log file path"`
	LogLevel string `flag:"level" desc:"log level: INFO/WARN/DEBUG"`
}

// DataAgingConfig 数据老化配置
type DataAgingConfig struct {
	ScanningTaskPeriod          string `flag:"scanning_task_period"`
	ClearingTaskPeriod          string `flag:"clearing_task_period"`
	ClearingTaskThreshold       string `flag:"clearing_task_threshold"`
	FileAccessInactiveThreshold string `flag:"file_access_inactive_threshold"`
	DeleteInactiveFileTimeout   string `flag:"delete_inactive_file_timeout"`
	CacheAvailable              bool
}

var (
	config *Config
	mu     sync.RWMutex
)

// GlobalConfig 全局配置实例（供 flag 解析直接引用）
var GlobalConfig *Config

func init() {
	config = &Config{
		MediaCache:  getEnv("MEDIA_CACHE_PATH", "/opt/mtuser/mcs/video"),
		HTTPTimeout: 600,
		Logger: LoggerConfig{
			LogFile:  getEnv("LOG_FILE", "/opt/mtuser/mcs/log/log1"),
			LogLevel: getEnv("LOG_LEVEL", "INFO"),
		},
		DataAging: DataAgingConfig{
			ScanningTaskPeriod:          getEnv("SCANNING_TASK_PERIOD", "1"),
			ClearingTaskPeriod:          getEnv("CLEARING_TASK_PERIOD", "24"),
			ClearingTaskThreshold:       getEnv("CLEARING_TASK_THRESHOLD", "491520"),
			FileAccessInactiveThreshold: getEnv("FILE_ACCESS_INACTIVE_THRESHOLD", "10"),
			DeleteInactiveFileTimeout:   getEnv("DELETE_INACTIVE_FILE_TIMEOUT", "60"),
			CacheAvailable:              false,
		},
	}
	GlobalConfig = config
}

// Instance 获取全局配置单例（线程安全）
func Instance() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

// SetCacheAvailable 设置缓存可用状态（线程安全）
func SetCacheAvailable(available bool) {
	mu.Lock()
	defer mu.Unlock()
	config.DataAging.CacheAvailable = available
}

// IsCacheAvailable 获取缓存可用状态（线程安全）
func IsCacheAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return config.DataAging.CacheAvailable
}

// Init 显式初始化入口，供 main 包调用（init() 已自动执行，此处为空占位）
func Init() {}

// getEnv 从环境变量获取值，不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
