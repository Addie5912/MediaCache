package flagutil

import (
	"os"
	"testing"
)

// TestConfig 测试用配置结构体
type TestConfig struct {
	Host    string `flag:"host" desc:"服务地址"`
	Port    int    `flag:"port" desc:"服务端口"`
	Debug   bool   `flag:"debug" desc:"调试模式"`
	Timeout float64 `flag:"timeout" desc:"超时时间(秒)"`
	Logger  LoggerConfig
}

// LoggerConfig 嵌套日志配置
type LoggerConfig struct {
	LogFile  string `flag:"file" desc:"日志文件路径"`
	LogLevel string `flag:"level" desc:"日志级别"`
}

// TestParse_DefaultValues 测试未传入命令行参数时使用默认值
func TestParse_DefaultValues(t *testing.T) {
	cfg := &TestConfig{
		Host:    "localhost",
		Port:    8080,
		Debug:   false,
		Timeout: 30.0,
		Logger: LoggerConfig{
			LogFile:  "/var/log/app.log",
			LogLevel: "INFO",
		},
	}

	// 清空命令行参数，模拟无参数启动
	os.Args = []string{"test"}

	result := Parse(cfg).(*TestConfig)

	if result.Host != "localhost" {
		t.Errorf("expected Host=localhost, got %s", result.Host)
	}
	if result.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", result.Port)
	}
	if result.Debug != false {
		t.Errorf("expected Debug=false, got %v", result.Debug)
	}
	if result.Timeout != 30.0 {
		t.Errorf("expected Timeout=30.0, got %f", result.Timeout)
	}
	if result.Logger.LogFile != "/var/log/app.log" {
		t.Errorf("expected LogFile=/var/log/app.log, got %s", result.Logger.LogFile)
	}
	if result.Logger.LogLevel != "INFO" {
		t.Errorf("expected LogLevel=INFO, got %s", result.Logger.LogLevel)
	}
}

// TestParse_WithArgs 测试传入命令行参数时覆盖默认值
func TestParse_WithArgs(t *testing.T) {
	cfg := &TestConfig{
		Host:    "localhost",
		Port:    8080,
		Debug:   false,
		Timeout: 30.0,
		Logger: LoggerConfig{
			LogFile:  "/var/log/app.log",
			LogLevel: "INFO",
		},
	}

	os.Args = []string{
		"test",
		"-host", "192.168.1.1",
		"-port", "9090",
		"-debug",
		"-timeout", "60.5",
		"-file", "/tmp/test.log",
		"-level", "DEBUG",
	}

	result := Parse(cfg).(*TestConfig)

	if result.Host != "192.168.1.1" {
		t.Errorf("expected Host=192.168.1.1, got %s", result.Host)
	}
	if result.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", result.Port)
	}
	if result.Debug != true {
		t.Errorf("expected Debug=true, got %v", result.Debug)
	}
	if result.Timeout != 60.5 {
		t.Errorf("expected Timeout=60.5, got %f", result.Timeout)
	}
	if result.Logger.LogFile != "/tmp/test.log" {
		t.Errorf("expected LogFile=/tmp/test.log, got %s", result.Logger.LogFile)
	}
	if result.Logger.LogLevel != "DEBUG" {
		t.Errorf("expected LogLevel=DEBUG, got %s", result.Logger.LogLevel)
	}
}
