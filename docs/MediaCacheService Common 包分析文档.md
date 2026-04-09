# MediaCacheService Common 包分析文档

## 目录
1. [概述](#概述)
2. [目录结构](#目录结构)
3. [接口分析](#接口分析)
4. [结构体分析](#结构体分析)
5. [函数实现详解](#函数实现详解)
6. [调用关系图](#调用关系图)
7. [设计模式和架构原则](#设计模式和架构原则)
8. [包间依赖关系](#包间依赖关系)

## 概述

MediaCacheService 的 `common` 包是整个系统的基础设施层，提供了配置管理、日志记录、HTTP/HTTPS 服务器、错误处理、常量定义等核心功能。该包采用模块化设计，各子包职责明确，为上层业务逻辑提供稳定可靠的基础服务。

### 技术栈
- **Go 1.20**
- **Beego v2** - Web 框架
- **Go-chassis/lager** - 日志框架
- **Go-chassis/GSF** - 企业服务治理框架
- **标准库**: net, crypto/tls, os, sync, time, encoding/json

### 核心特性
- 线程安全的配置管理
- 支持 HTTP 和 HTTPS 双协议服务器
- 完善的 TLS 证书管理和验证
- 多级日志和审计日志支持
- 灵活的错误处理机制
- 网络接口 IP 地址自动发现

## 目录结构

```
src/common/
├── conf/                    # 配置管理模块
│   └── config.go           # 应用配置实现
├── constants/              # 常量定义模块
│   ├── base.go             # 基础常量
│   └── retcode/            # 返回码常量
│       └── retcode.go      # API 返回码定义
├── error/                  # 错误处理模块
│   └── error.go            # 自定义错误类型
├── https/                  # HTTP/HTTPS 服务器模块
│   ├── http_server.go      # HTTP 服务器实现
│   ├── https_server.go     # HTTPS 服务器实现
│   ├── tls.go              # TLS 配置和证书验证
│   ├── http_server_test.go # HTTP 服务器测试
│   └── tls_test.go         # TLS 测试
└── logger/                 # 日志模块
    ├── logger.go           # 基础日志功能
    └── auditlog.go         # 审计日志功能
```

### 各模块职责

#### conf/ - 配置管理
- 集中管理应用配置
- 支持环境变量配置
- 提供线程安全的配置访问
- 支持配置的热更新

#### constants/ - 常量定义
- 定义服务相关常量
- 标准 API 返回码
- 环境变量名称定义

#### error/ - 错误处理
- 自定义错误类型
- 统一的错误处理接口
- 特定场景的错误判断

#### https/ - 服务器基础设施
- HTTP 服务器实现
- HTTPS 服务器实现
- TLS 证书管理
- IP 地址自动发现
- 过滤器管理

#### logger/ - 日志系统
- 基础日志封装
- 审计日志记录
- 远程日志提交
- 多级日志支持

## 接口分析

### 1. BeegoServer 接口

**定义位置**: `src/common/https/http_server.go`

```go
type BeegoServer interface {
    Run()
    Router(rootpath string, c beego.ControllerInterface, mappingMethods ...string) BeegoServer
    InsertFilter(pattern string, pos int, filter beego.FilterFunc, opts ...beego.FilterOpt) BeegoServer
}
```

**功能说明**:
- 定义服务器的基本操作契约
- 支持 HTTP 和 HTTPS 服务器的统一抽象
- 提供方法链式调用支持

**方法详解**:

#### 1.1 Run()
- **功能**: 启动服务器
- **返回**: 无
- **说明**: 异步启动服务器，通常在 goroutine 中调用

#### 1.2 Router()
- **功能**: 注册路由到控制器
- **参数**:
  - `rootpath`: URL 路径前缀
  - `c`: Beego 控制器接口
  - `mappingMethods`: 路由方法映射（可选）
- **返回**: BeegoServer 接口，支持链式调用
- **示例**: `server.Router("/video", &VideoController{}, "GET:GetVideo")`

#### 1.3 InsertFilter()
- **功能**: 插入过滤器（中间件）
- **参数**:
  - `pattern`: 匹配模式
  - `pos`: 过滤器执行位置（BeforeRouter, BeforeExec, AfterExec）
  - `filter`: 过滤器函数
  - `opts`: 可选配置
- **返回**: BeegoServer 接口，支持链式调用
- **示例**: `server.InsertFilter("*", beego.BeforeRouter, OverLoadFilter)`

**实现类**:
- `BeegoHttpServer` - HTTP 服务器实现
- `BeegoHttpsServer` - HTTPS 服务器实现

## 结构体分析

### 1. 配置相关结构体

#### 1.1 Config 结构体

**定义位置**: `src/common/conf/config.go`

```go
type Config struct {
    Logger      LoggerConfig `flag:"log"`
    MediaCache  string       `flag:"cache" desc:"video local cache address"`
    HTTPTimeout int
    DataAging   DataAgingConfig `flag:"data"`
}
```

**字段说明**:
- `Logger`: 日志配置
- `MediaCache`: 媒体缓存目录路径
- `HTTPTimeout`: HTTP 请求超时时间（秒）
- `DataAging`: 数据老化配置

**设计特点**:
- 使用 tag 支持命令行参数解析
- 嵌套结构体组织相关配置
- 支持多层级配置管理

#### 1.2 LoggerConfig 结构体

```go
type LoggerConfig struct {
    LogFile  string `flag:"file" desc:"log file path"`
    LogLevel string `flag:"level" desc:"log level: INFO/WARN/DEBUG"`
}
```

**字段说明**:
- `LogFile`: 日志文件路径，默认 `/opt/mtuser/mcs/log/log1`
- `LogLevel`: 日志级别，支持 INFO/WARN/DEBUG，默认 INFO

**使用示例**:
```go
logger := conf.Instance().Logger
fmt.Printf("Log file: %s, Level: %s", logger.LogFile, logger.LogLevel)
```

#### 1.3 DataAgingConfig 结构体

```go
type DataAgingConfig struct {
    ScanningTaskPeriod          string `flag:"scanning_task_period"`
    ClearingTaskPeriod          string `flag:"clearing_task_period"`
    ClearingTaskThreshold       string `flag:"clearing_task_threshold"`
    FileAccessInactiveThreshold string `flag:"file_access_inactive_threshold"`
    DeleteInactiveFileTimeout   string `flag:"delete_inactive_file_timeout"`
    CacheAvailable              bool
}
```

**字段详解**:

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `ScanningTaskPeriod` | string | "1" | 扫描任务周期（小时） |
| `ClearingTaskPeriod` | string | "24" | 清理任务周期（小时） |
| `ClearingTaskThreshold` | string | "491520" | 触发清理的磁盘阈值（MB，约480GB） |
| `FileAccessInactiveThreshold` | string | "10" | 文件未活跃阈值（天） |
| `DeleteInactiveFileTimeout` | string | "60" | 删除未活跃文件超时时间（秒） |
| `CacheAvailable` | bool | false | 缓存是否可用标志 |

**数据老化机制**:
1. 定期扫描缓存文件状态
2. 根据磁盘空间使用情况决定是否清理
3. 清理超过阈值未访问的文件
4. 支持动态控制缓存可用性

### 2. 服务器结构体

#### 2.1 BeegoHttpServer 结构体

**定义位置**: `src/common/https/http_server.go`

```go
type BeegoHttpServer struct {
    server *beego.HttpServer
    ip     string
    port   int
}
```

**字段说明**:
- `server`: Beego HTTP 服务器实例
- `ip`: 监听的 IP 地址
- `port`: 监听的端口号

**实现接口**: BeegoServer

**主要功能**:
- 封装 Beego HTTP 服务器
- 管理 IP 和端口配置
- 提供路由和过滤器注册功能

#### 2.2 BeegoHttpsServer 结构体

**定义位置**: `src/common/https/https_server.go`

```go
type BeegoHttpsServer struct {
    server        *beego.HttpServer
    ip            string
    port          int
    certInfo      CertInfo
    restartChan   chan CertInfo
    isServerReady bool
}
```

**字段说明**:
- `server`: Beego HTTPS 服务器实例
- `ip`: 监听的 IP 地址
- `port`: 监听的端口号
- `certInfo`: TLS 证书信息
- `restartChan`: 证书更新通道（用于热更新）
- `isServerReady`: 服务器就绪标志

**实现接口**: BeegoServer

**增强功能**:
- 动态证书管理
- 证书热更新支持
- 安全启动机制
- 服务器生命周期管理

### 3. TLS 相关结构体

#### 3.1 CertInfo 结构体

**定义位置**: `src/common/https/tls.go`

```go
type CertInfo struct {
    KeyFile  string
    CertFile string
    CaFile   string
    KeyPwd   []byte
}
```

**字段说明**:
- `KeyFile`: 私钥文件路径
- `CertFile`: 证书文件路径
- `CaFile`: CA 证书文件路径
- `KeyPwd`: 私钥密码（支持加密私钥）

**证书格式支持**:
- PEM 格式证书
- 加密私钥（PKCS#8）
- 自动密码解密

#### 3.2 verify 结构体

```go
type verify struct {
    rootCA *x509.CertPool
}
```

**字段说明**:
- `rootCA`: 根证书池，用于证书链验证

**功能**:
- 证书有效性验证
- 密钥用法检查
- 签名算法验证
- 主机名匹配
- 证书链验证

### 4. 日志相关结构体

#### 4.1 AuditsInfo 结构体

**定义位置**: `src/common/logger/auditlog.go`

```go
type AuditsInfo struct {
    Terminal string
    UserName string
    Detail   string
    DetailZh string
}
```

**字段说明**:
- `Terminal`: 终端标识
- `UserName`: 操作用户名
- `Detail`: 操作详情（英文）
- `DetailZh`: 操作详情（中文）

#### 4.2 AuditsPara 结构体

```go
type AuditsPara struct {
    OperationZH string
    OperationEN string
    OperateType OperateType
    Level       AuditLogLevel
    Username    string
    Terminal    string
    Result      int
    Detail      string
    DetailZH    string
}
```

**字段说明**:
- `OperationZH`: 操作名称（中文）
- `OperationEN`: 操作名称（英文）
- `OperateType`: 操作类型（GET/ADD/MOD/DELETE/DOWNLOAD/UPLOAD/UPHOLD）
- `Level`: 操作日志级别（Minor/Important/Auto/Manual）
- `Username`: 操作用户
- `Terminal`: 操作终端
- `Result`: 操作结果
- `Detail`: 操作详情（英文）
- `DetailZH`: 操作详情（中文）

**操作类型枚举**:
```go
type OperateType int

const (
    GET      OperateType = iota  // 读取操作
    ADD                           // 新增操作
    MOD                           // 修改操作
    DELETE                        // 删除操作
    DOWNLOAD                      // 下载操作
    UPLOAD                        // 上传操作
    UPHOLD                        // 维护操作
)
```

**日志级别枚举**:
```go
type AuditLogLevel int

const (
    MinorLevel     AuditLogLevel = 0  // 次要操作
    ImportantLevel AuditLogLevel = 1  // 重要操作
    LogLevelAuto   AuditLogLevel = 3  // 自动查询
    LogLevelManual AuditLogLevel = 4  // 手动查询
)
```

### 5. 错误处理结构体

#### 5.1 Err 类型

**定义位置**: `src/common/error/error.go`

```go
type Err string
```

**功能**:
- 自定义错误类型
- 实现标准 error 接口
- 用于特定场景的错误判断

**常量定义**:
```go
const (
    ErrNotExist Err = "data not exist"
)
```

**使用示例**:
```go
func getData(key string) (string, error) {
    if !exists(key) {
        return "", error.ErrNotExist
    }
    return data[key], nil
}

// 检查特定错误
if error.IsNotExist(err) {
    // 处理数据不存在的情况
}
```

## 函数实现详解

### 1. 配置管理函数

#### 1.1 Instance() 函数

**定义位置**: `src/common/conf/config.go`

```go
func Instance() *Config
```

**功能描述**:
- 获取全局配置的单例实例
- 使用读写锁保证线程安全

**实现逻辑**:
1. 获取读锁
2. 返回全局配置实例
3. 释放读锁

**线程安全性**:
- 使用 RWMutex 保护配置访问
- 读操作可以并发执行
- 写操作（SetCacheAvailable）使用写锁独占访问

**使用示例**:
```go
config := conf.Instance()
logLevel := config.Logger.LogLevel
cachePath := config.MediaCache
```

#### 1.2 SetCacheAvailable() 函数

```go
func SetCacheAvailable(available bool)
```

**功能描述**:
- 设置缓存可用性状态
- 用于动态控制缓存功能

**实现逻辑**:
1. 获取写锁
2. 更新缓存可用状态
3. 释放写锁

**线程安全性**:
- 使用写锁确保原子性操作
- 防止并发修改导致的不一致

**使用场景**:
- 磁盘空间不足时禁用缓存
- 维护模式下禁用缓存
- 根据系统负载动态调整

#### 1.3 IsCacheAvailable() 函数

```go
func IsCacheAvailable() bool
```

**功能描述**:
- 检查缓存是否可用
- 便捷函数，直接返回缓存状态

**使用示例**:
```go
if !conf.IsCacheAvailable() {
    return "缓存服务暂时不可用"
}
```

#### 1.4 getEnv() 函数

```go
func getEnv(key string, defaultVal string) string
```

**功能描述**:
- 获取环境变量，支持默认值
- 用于配置的灵活初始化

**实现逻辑**:
1. 尝试读取环境变量
2. 如果环境变量未设置，返回默认值
3. 否则返回环境变量的值

**使用示例**:
```go
// 在 init() 函数中使用
ScanningTaskPeriod: getEnv("SCANNING_TASK_PERIOD", "1"),
ClearingTaskPeriod: getEnv("CLEARING_TASK_PERIOD", "24"),
```

### 2. 服务器创建函数

#### 2.1 NewHttpServer() 函数

**定义位置**: `src/common/https/http_server.go`

```go
func NewHttpServer(ip string, port int) *BeegoHttpServer
```

**功能描述**:
- 创建 HTTP 服务器实例
- 基于 Beego 框架

**实现逻辑**:
1. 复制 Beego 默认配置
2. 创建新的 HTTP 服务器
3. 设置监听 IP 和端口
4. 返回服务器实例

**参数说明**:
- `ip`: 监听的 IP 地址（"0.0.0.0" 表示监听所有网卡）
- `port`: 监听的端口号

**使用示例**:
```go
server := https.NewHttpServer("0.0.0.0", 8080)
server.Router("/test", &TestController{}, "GET:Test")
server.Run()
```

#### 2.2 NewHttpsServer() 函数

**定义位置**: `src/common/https/https_server.go`

```go
func NewHttpsServer(ip string, port int) *BeegoHttpsServer
```

**功能描述**:
- 创建 HTTPS 服务器实例
- 支持动态证书管理

**实现逻辑**:
1. 创建 Beego HTTPS 服务器配置
2. 初始化证书更新通道
3. 设置服务器未就绪标志
4. 返回服务器实例

**增强功能**:
- 证书自动更新机制
- 安全启动（等待证书上传）
- 服务器热重启支持

**使用示例**:
```go
server := https.NewHttpsServer("0.0.0.0", 8443)
server.Router("/secure", &SecureController{}, "GET:GetData")
server.RunWithPresetCert() // 使用预设证书启动
```

#### 2.3 newBeegoHttpsServer() 函数

```go
func newBeegoHttpsServer(ip string, port int) *beego.HttpServer
```

**功能描述**:
- 创建底层 Beego HTTPS 服务器
- 配置 TLS 相关设置

**实现逻辑**:
1. 复制 Beego 默认配置
2. 启用 HTTPS
3. 设置监听 IP 和端口
4. 禁用 HTTP 监听

**安全配置**:
- 强制使用 HTTPS
- TLS 配置在 GetTLS() 中设置

### 3. IP 地址发现函数

#### 3.1 GetLocalIP() 函数

**定义位置**: `src/common/https/http_server.go`

```go
func GetLocalIP(ethEnv, defaultEth string) (string, error)
```

**功能描述**:
- 从环境变量或默认网卡获取 IP 地址
- 支持多网卡环境下的精确选择

**实现逻辑**:
1. 从环境变量读取网卡名称
2. 如果环境变量未设置，使用默认网卡名
3. 调用 getEthIP() 获取指定网卡的 IP
4. 返回 IPv4 地址或错误

**参数说明**:
- `ethEnv`: 网卡名称的环境变量名（如 "FABRIC_ETH"）
- `defaultEth`: 默认网卡名称（如 "bond-base"）

**使用示例**:
```go
// 内部网络
internalIP, err := https.GetLocalIP("FABRIC_ETH", "bond-base")

// 外部网络
externalIP, err := https.GetLocalIP("SC_TRUNK_ETH", "bond-external")
```

**错误处理**:
- 网卡不存在
- 网卡没有配置 IP
- IP 地址获取失败

#### 3.2 getEthIP() 函数

```go
func getEthIP(ethName string) (string, error)
```

**功能描述**:
- 获取指定网卡的 IPv4 地址
- 内部函数，不对外暴露

**实现逻辑**:
1. 通过网卡名称获取网络接口
2. 获取接口的所有地址
3. 遍历地址，查找 IPv4 地址
4. 返回第一个 IPv4 地址

**地址类型判断**:
```go
switch ip := addr.(type) {
case *net.IPNet:
    if ip.IP.To4() != nil {
        // IPv4 地址
        return ip.IP.String(), nil
    }
}
```

**限制**:
- 只返回 IPv4 地址
- 如果有多个 IPv4 地址，返回第一个

### 4. TLS 相关函数

#### 4.1 GetTLS() 函数

**定义位置**: `src/common/https/tls.go`

```go
func GetTLS(info CertInfo, tlsType string) *tls.Config
```

**功能描述**:
- 生成 TLS 配置
- 支持服务端和客户端配置

**实现逻辑**:
1. 创建基础 TLS 配置
2. 设置支持的协议版本（TLS 1.2-1.3）
3. 配置安全的加密套件
4. 加载证书和 CA
5. 根据类型配置服务端或客户端参数

**安全配置**:

**协议版本**:
- MinVersion: TLS 1.2
- MaxVersion: TLS 1.3

**支持的加密套件**:
```go
tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
```

**ALPN 支持**:
```go
NextProtos: []string{"h2", "http/1.1"},
```

**服务端配置**:
- 设置 ClientCAs 用于客户端证书验证

**客户端配置**:
- 设置 RootCAs 用于服务器证书验证
- 自定义证书验证函数
- InsecureSkipVerify 设置为 true，使用自定义验证

**使用示例**:
```go
// 服务端
tlsConfig := GetTLS(certInfo, ServerType)
server.TLSConfig = tlsConfig

// 客户端
tlsConfig := GetTLS(certInfo, ClientType)
client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
```

#### 4.2 getCert() 函数

```go
func getCert(info CertInfo) (*x509.CertPool, *tls.Certificate, error)
```

**功能描述**:
- 加载证书文件并生成证书对
- 支持加密私钥

**实现逻辑**:
1. 读取证书文件
2. 读取私钥文件
3. 如果有密码，解密私钥（PKCS#8）
4. 生成证书对
5. 读取 CA 证书并创建证书池
6. 返回证书池和证书对

**私钥解密**:
```go
if len(info.KeyPwd) != 0 {
    privBlock, _ := pem.Decode(tlsKey)
    decryptedBlock, err := x509.DecryptPEMBlock(privBlock, info.KeyPwd)
    tlsKey = pem.EncodeToMemory(&pem.Block{
        Type:  privBlock.Type,
        Bytes: decryptedBlock,
    })
}
```

**错误处理**:
- 文件读取失败
- 私钥解密失败
- 证书对生成失败
- CA 证书加载失败

#### 4.3 defaultLoadFile() 函数

```go
func defaultLoadFile(filePath string) ([]byte, error)
```

**功能描述**:
- 安全加载文件内容
- 转换为绝对路径

**实现逻辑**:
1. 将路径转换为绝对路径
2. 使用 os.ReadFile 读取文件
3. 返回文件内容

**安全性**:
- 使用绝对路径防止路径遍历
- 使用标准库读取文件

### 5. BeegoHttpsServer 方法

#### 5.1 Run() 方法

**定义位置**: `src/common/https/https_server.go`

```go
func (b *BeegoHttpsServer) Run()
```

**功能描述**:
- 异步启动 HTTPS 服务器
- 实现证书动态更新机制

**实现逻辑**:
1. 启动 goroutine 监听证书更新通道
2. 等待证书上传
3. 收到证书后更新 TLS 配置
4. 启动服务器
5. 标记服务器为就绪状态

**证书更新流程**:
```
等待证书 -> 收到证书 -> 检测服务器状态 ->
├─ 未启动 -> 启动服务器
└─ 已启动 -> 退出进程（触发重启）
```

**安全机制**:
- 必须收到完整证书才启动
- 证书更新触发安全重启
- 服务器就绪标志防止重复启动

#### 5.2 RunWithPresetCert() 方法

```go
func (b *BeegoHttpsServer) RunWithPresetCert()
```

**功能描述**:
- 使用预设证书启动服务器
- 直接启动，无需等待证书上传

**实现逻辑**:
1. 从环境变量读取证书路径
2. 加载证书文件
3. 配置 TLS
4. 启动服务器

**环境变量**:
- `SSLPATH`: SSL 证书目录路径
- `INNER_TLS_PRIVATE_KEY_PWD`: 私钥密码

**证书文件结构**:
```
$SSLPATH/
├── ca.crt
├── tls.crt
└── tls.key.pwd
```

#### 5.3 UpdateCert() 方法

```go
func (b *BeegoHttpsServer) UpdateCert(certInfo CertInfo)
```

**功能描述**:
- 更新服务器证书
- 触发证书更新流程

**实现逻辑**:
1. 将证书信息发送到更新通道
2. 的 goroutine 接收并处理
3. 根据服务器状态决定重启或启动

**证书更新时机**:
- 证书即将过期
- CA 证书更新
- 密钥轮换

#### 5.4 updateCertInfo() 方法

```go
func (b *BeegoHttpsServer) updateCertInfo(certInfo CertInfo)
```

**功能描述**:
- 更新内部证书信息
- 分字段更新，避免覆盖

**实现逻辑**:
1. 如果有 KeyFile，更新密钥文件和密码
2. 如果有 CaFile，更新 CA 文件

#### 5.5 needStartServer() 方法

```go
func (b *BeegoHttpsServer) needStartServer() bool
```

**功能描述**:
- 判断是否需要启动服务器
- 处理证书更新场景

**决策逻辑**:
```
证书完整 + 服务器未启动 -> 启动
证书完整 + 服务器已启动 -> 重启（退出码 3）
证书不完整 -> 跳过
```

#### 5.6 close() 方法

```go
func (b *BeegoHttpsServer) close()
```

**功能描述**:
- 关闭 HTTPS 服务器
- 内部方法

### 6. 证书验证函数

#### 6.1 verify.verifyConnection() 方法

```go
func (v *verify) verifyConnection(cs tls.ConnectionState) error
```

**功能描述**:
- 综合验证 TLS 连接
- 调用多个子验证函数

**验证流程**:
1. 检查对端证书是否存在
2. 验证证书有效性（有效期）
3. 验证基本约束（不能是 CA）
4. 验证签名算法（不允许弱算法）
5. 验证密钥用法
6. 验证主机名或 IP 地址

**实现**:
```go
cert := cs.PeerCertificates[0]

if err := v.checkValidity(cert); err != nil {
    return err
}
if err := v.checkBasicConstraints(cert); err != nil {
    return err
}
if err := v.checkSignatureAlgorithm(cert); err != nil {
    return err
}
if err := v.checkKeyUsage(cert); err != nil {
    return err
}

if len(cert.DNSNames) > 0 {
    if err := v.verifyHostname(cert, cs.ServerName); err != nil {
        return err
    }
}

return nil
```

#### 6.2 checkValidity() 方法

```go
func (v *verify) checkValidity(cert *x509.Certificate) error
```

**功能描述**:
- 验证证书有效期
- 检查是否在有效期内

**验证条件**:
- 未过期：`now.After(cert.NotAfter)` 失败
- 未提前生效：`now.Before(cert.NotBefore)` 失败

#### 6.3 checkBasicConstraints() 方法

```go
func (v *verify) checkBasicConstraints(cert *x509.Certificate) error
```

**功能描述**:
- 验证基本约束
- 确保服务器证书不是 CA 证书

**安全检查**:
```go
if cert.IsCA {
    return fmt.Errorf("server certificate cannot be a CA certificate")
}
```

#### 6.4 checkSignatureAlgorithm() 方法

```go
func (v *verify) checkSignatureAlgorithm(cert *x509.Certificate) error
```

**功能描述**:
- 验证签名算法安全性
- 拒绝弱签名算法

**弱算法列表**:
```go
weakAlgorithms := map[x509.SignatureAlgorithm]bool{
    x509.MD2WithRSA:    true,
    x509.MD5WithRSA:    true,
    x509.DSAWithSHA1:   true,
    x509.ECDSAWithSHA1: true,
    x509.SHA1WithRSA:   true,
}
```

#### 6.5 checkKeyUsage() 方法

```go
func (v *verify) checkKeyUsage(cert *x509.Certificate) error
```

**功能描述**:
- 验证密钥用法
- 拒绝不安全的密钥用法

**安全检查**:
```go
if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
    return fmt.Errorf("cert should not allow certificate signing")
}

if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
    return fmt.Errorf("certificate should not allow CRL signature")
}
```

#### 6.6 verifyHostname() 方法

```go
func (v *verify) verifyHostname(cert *x509.Certificate, hostname string) error
```

**功能描述**:
- 验证主机名或 IP 地址
- 支持通配符证书

**验证流程**:
1. 首先检查 SAN (Subject Alternative Names) 中的 DNS 名称
2. 检查 IP 地址（如果 hostname 是 IP）
3. 使用 matchHostname 进行匹配

**通配符匹配**:
- `*.example.com` 匹配 `foo.example.com`
- 不匹配 `bar.foo.example.com`

#### 6.7 matchHostname() 函数

```go
func matchHostname(pattern, hostname string) bool
```

**功能描述**:
- 主机名匹配函数
- 支持通配符

**匹配规则**:
1. 统一转换为小写
2. 去除末尾点号
3. 按点号分割
4. `*` 匹配任意部分
5. 部分数量必须相等

**示例**:
```go
matchHostname("*.example.com", "foo.example.com")  // true
matchHostname("*.example.com", "foo.bar.example.com")  // false
matchHostname("foo.example.com", "foo.example.com")  // true
```

### 7. 日志函数

#### 7.1 Infof() 函数

**定义位置**: `src/common/logger/logger.go`

```go
func Infof(format string, args ...interface{})
```

**功能描述**:
- 记录信息级别日志
- 封装 Go-chassis/lager 框架

**使用示例**:
```go
logger.Infof("Server started on port %d", port)
logger.Infof("Request received from %s", clientIP)
```

#### 7.2 Warnf() 函数

```go
func Warnf(format string, args ...interface{})
```

**功能描述**:
- 记录警告级别日志
- 用于非关键性问题

**使用示例**:
```go
logger.Warnf("Cache miss for key %s", key)
logger.Warnf("Slow query detected: %v elapsed", duration)
```

#### 7.3 Debugf() 函数

```go
func Debugf(format string, args ...interface{})
```

**功能描述**:
- 记录调试级别日志
- 用于开发调试

**使用示例**:
```go
logger.Debugf("Processing request: %+v", request)
logger.Debugf("Cache lookup result: %v", found)
```

#### 7.4 Errorf() 函数

```go
func Errorf(format string, args ...interface{})
```

**功能描述**:
- 记录错误级别日志
- 用于错误情况

**使用示例**:
```go
logger.Errorf("Failed to connect to server: %v", err)
logger.Errorf("Invalid request parameters: %v", params)
```

#### 7.5 TeeErrorf() 函数

```go
func TeeErrorf(format string, args ...interface{}) error
```

**功能描述**:
- 记录错误日志并返回错误对象
- 便捷的错误处理

**使用示例**:
```go
if err != nil {
    return logger.TeeErrorf("Failed to process: %v", err)
}
```

#### 7.6 Fatalf() 函数

```go
func Fatalf(format string, args ...interface{})
```

**功能描述**:
- 记录致命错误日志并退出程序
- 用于不可恢复的错误

**使用示例**:
```go
logger.Fatalf("Fatal error: %v", err)
// 程序将在此退出
```

### 8. 审计日志函数

#### 8.1 AuditsLog() 函数

**定义位置**: `src/common/logger/auditlog.go`

```go
func AuditsLog(auditsPara *AuditsPara, requestURL string)
```

**功能描述**:
- 提交审计日志到远程服务
- 支持操作日志和安全日志

**实现逻辑**:
1. 构建日志请求体
2. 序列化为 JSON（双重序列化）
3. 调用 GSF API 提交日志
4. 处理响应

**请求 URL**:
- 操作日志: `cse://AuditLog/plat/audit/v1/logs`
- 安全日志: `cse://AuditLog/plat/audit/v1/seculogs`

**日志体结构**:
```json
{
  "operation": "{\"OP_ZH\":\"操作\",\"OP_EN\":\"operation\"}",
  "level": 0,
  "userName": "admin",
  "dateTime": 1234567890,
  "appName": "mediacache",
  "appId": "mediacache",
  "terminal": "web",
  "serviceName": "MCS",
  "result": 0,
  "detail": "operation detail",
  "detail_zh": "操作详情"
}
```

**双重序列化**:
```go
// 第一次序列化
bs, err := json.Marshal(body)

// 第二次序列化为字符串
bs2, err := json.Marshal(string(bs))
```

**为什么双重序列化**:
- 审计日志服务端要求接收字符串
- 字符串本身是 JSON 格式
- 确保数据传输格式正确

**使用示例**:
```go
auditsPara := &logger.AuditsPara{
    OperationZH: "访问视频",
    OperationEN: "Access Video",
    OperateType: logger.GET,
    Level:       logger.MinorLevel,
    Username:    "user123",
    Terminal:    "mobile",
    Result:      0,
    Detail:      "Access video file sample.mp4",
    DetailZH:    "访问视频文件 sample.mp4",
}

logger.AuditsLog(auditsPara, logger.OpsLog)
```

#### 8.2 AuditsSecAndOpsLog() 函数

```go
func AuditsSecAndOpsLog(secAuditsPara, opsAuditsPara *AuditsPara)
```

**功能描述**:
- 同时记录安全日志和操作日志
- 用于需要双重审计的场景

**实现逻辑**:
```go
AuditsLog(secAuditsPara, SecLog)
AuditsLog(opsAuditsPara, OpsLog)
```

**使用场景**:
- 敏感操作需要同时记录到安全日志和操作日志
- 安全审计和运维审计分离管理

## 调用关系图

### 1. 整体架构调用关系

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│                   (controllers, services)                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                      Common Package                          │
│                                                               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │ conf/      │  │ https/     │  │ logger/    │            │
│  │            │  │            │  │            │            │
│  │ Config     │  │ BeegoServer│  │ Logger     │            │
│  │ DataAging  │  │ TLS Config │  │ AuditLog   │            │
│  └────────────┘  └────────────┘  └────────────┘            │
│                              │                               │
│  ┌────────────┐              │        ┌────────────┐        │
│  │ constants/ │              ├───────►│ error/     │        │
│  │            │              │        │            │        │
│  │ retcode    │              │        │ Err Type   │        │
│  └────────────┘              │        └────────────┘        │
│                             ▼                               │
│              ┌───────────────────────┐                      │
│              │   External Systems    │                      │
│              │                       │                      │
│              │  - GSF API (AuditLog) │                      │
│              │  - CSE Audit Service  │                      │
│              └───────────────────────┘                      │
└─────────────────────────────────────────────────────────────┘
```

### 2. 配置管理调用链

```
main()
├── conf.Instance()
│   ├── 读锁获取
│   ├── 返回全局配置实例
│   └── 释放读锁
├── conf.SetCacheAvailable()
│   ├── 获取写锁
│   ├── 更新缓存状态
│   └── 释放写锁
└── conf.IsCacheAvailable()
    └── 返回缓存状态

init()
└── 初始化默认配置
    ├── conf.getEnv() 读取环境变量
    ├── 设置默认日志配置
    ├── 设置缓存路径
    ├── 设置数据老化配置
    └── 完成配置初始化
```

### 3. 服务器启动调用链

```
main()
├── startInternalServer()
│   ├── https.GetLocalIP("FABRIC_ETH", "bond-base")
│   │   ├── os.Getenv(ethEnv)
│   │   ├── https.getEthIP(netif)
│   │   │   ├── net.InterfaceByName()
│   │   │   ├── iface.Addrs()
│   │   │   └── 提取 IPv4 地址
│   │   └── 返回 IP 地址
│   ├── https.NewHttpServer(ip, port)
│   │   ├── 复制 Beego 配置
│   │   ├── 创建 HttpServer
│   │   └── 返回 BeegoHttpServer 实例
│   ├── routers.RegisterRouters(server)
│   │   ├── server.InsertFilter() 注册过滤器
│   │   └── server.Router() 注册路由
│   └── server.Run() 启动服务器
│
└── startExternalServer()
    ├── https.GetLocalIP("SC_TRUNK_ETH", "bond-external")
    ├── https.NewHttpsServer(ip, port)
    │   ├── newBeegoHttpsServer(ip, port)
    │   │   ├── 创建 HTTPS 服务器配置
    │   │   ├── 设置 TLS 模式
    │   │   └── 返回基础服务器
    │   ├── 初始化证书更新通道
    │   └── 返回 BeegoHttpsServer 实例
    ├── routers.RegisterRouters(server)
    └── server.RunWithPresetCert()
        ├── os.Getenv("SSLPATH")
        ├── 加载证书文件
        ├── https.GetTLS(certInfo, ServerType)
        │   ├── 创建基础 TLS 配置
        │   ├── https.getCert() 加载证书
        │   │   ├── defaultLoadFile() 读取文件
        │   │   ├── 解密私钥（如需）
        │   │   ├── 加载 CA 证书
        │   │   └── 生成证书对
        │   └── 配置服务端 TLS 参数
        └── server.Run() 启动 HTTPS 服务器
```

### 4. TLS 配置调用链

```
BeegoHttpsServer.RunWithPresetCert()
├── 构建证书信息
│   ├── CertFile: path.Join(sslPath, "tls.crt")
│   ├── KeyFile: path.Join(sslPath, "tls.key.pwd")
│   ├── CaFile: path.Join(sslPath, "ca.crt")
│   └── KeyPwd: os.Getenv("INNER_TLS_PRIVATE_KEY_PWD")
├── https.GetTLS(certInfo, ServerType)
│   ├── 创建基础 TLS 配置
│   │   ├── 设置协议版本 (TLS 1.2-1.3)
│   │   ├── 配置加密套件
│   │   └── 设置 ALPN ("h2", "http/1.1")
│   ├── https.getCert(certInfo)
│   │   ├── defaultLoadFile(certFile)
│   │   ├── defaultLoadFile(keyFile)
│   │   ├── 解密私钥（如果 KeyPwd 存在）
│   │   │   ├── pem.Decode()
│   │   │   ├── x509.DecryptPEMBlock()
│   │   │   └── pem.EncodeToMemory()
│   │   ├── tls.X509KeyPair()
│   │   ├── defaultLoadFile(caFile)
│   │   └── 创建 x509.CertPool
│   └── 配置服务端参数
│       └── tlsConfig.ClientCAs = caPool
└── 启动 TLS 服务器
```

### 5. 日志调用链

```
应用程序代码
├── logger.Infof()
│   └── lager.Logger.Infof()
├── logger.Errorf()
│   └── lager.Logger.Errorf()
└── logger.AuditsLog()
    ├── 构建审计日志参数 AuditsPara
    ├── 构建日志体
    │   ├── 序列化操作信息
    │   ├── 设置时间戳
    │   ├── 设置应用信息
    │   └── 详细信息
    ├── 序列化为 JSON (第一次)
    ├── 序列化为字符串 (第二次)
    ├── gsfapi.NewCspRestInvoker().Invoke()
    │   ├── POST 请求到审计日志服务
    │   ├── 发送序列化数据
    │   └── 接收响应
    └── 处理响应结果
        ├── 成功: 记录成功日志
        └── 失败: 记录错误日志
```

### 6. 证书验证调用链

```
TLS 连接建立
├── 服务器发送证书
├── 客户端验证证书
│   └── verify.verifyConnection()
│       ├── 检查对端证书是否存在
│       ├── verify.checkValidity(cert)
│       │   ├── 检查是否已生效
│       │   └── 检查是否已过期
│       ├── verify.checkBasicConstraints(cert)
│       │   └── 确保不是 CA 证书
│       ├── verify.checkSignatureAlgorithm(cert)
│       │   └── 拒绝弱算法 (MD2, MD5, SHA1)
│       ├── verify.checkKeyUsage(cert)
│       │   ├── 拒绝证书签名权限
│       │   └── 拒绝 CRL 签名权限
│       ├── verify.verifyHostname(cert, hostname)
│       │   ├── 检查 SAN 中的 DNS 名称
│       │   │   └── matchHostname()
│       │   │       ├── 转换为小写
│       │   │       ├── 分割域名部分
│       │   │       └── 通配符匹配
│       │   └── 检查 SAN 中的 IP 地址
│       │       └── IP 地址比较
│       └── 返回验证结果
└── 连接成功或失败
```

### 7. IP 地址发现调用链

```
main() -> startInternalServer() / startExternalServer()
├── https.GetLocalIP(ethEnv, defaultEth)
│   ├── eth := os.Getenv(ethEnv)
│   ├── if eth == "" { eth = defaultEth }
│   ├── https.getEthIP(eth)
│   │   ├── net.InterfaceByName(eth)
│   │   │   └── 获取网络接口
│   │   ├── iface.Addrs()
│   │   │   └── 获取接口地址
│   │   ├── 遍历地址
│   │   │   ├── case *net.IPNet
│   │   │   │   ├── if ip.IP.To4() != nil
│   │   │   │   └── return ip.IP.String()
│   │   │   └── default: 跳过
│   │   └── 错误处理
│   ├── if len(localIP) == 0 {
│   │   └── logger.Errorf("no ip on %v", eth)
│   │   └── return "", error
│   │   }
│   └── return localIP, nil
└── 使用 IP 地址启动服务器
    ├── server.Cfg.Listen.HTTPAddr = ip
    └── server.Run(ip, port)
```

## 设计模式和架构原则

### 1. 设计模式应用

#### 1.1 单例模式 (Singleton Pattern)

**应用场景**: 配置管理

**实现**:
```go
var (
    config *Config
    mu     sync.RWMutex
)

func Instance() *Config {
    mu.RLock()
    defer mu.RUnlock()
    return config
}
```

**特点**:
- 全局唯一配置实例
- 线程安全访问
- 延迟初始化

#### 1.2 策略模式 (Strategy Pattern)

**应用场景**: HTTP/HTTPS 服务器实现

**接口定义**:
```go
type BeegoServer interface {
    Run()
    Router(...)
    InsertFilter(...)
}

// 策略实现
type BeegoHttpServer struct { /* HTTP 实现 */ }
type BeegoHttpsServer struct { /* HTTPS 实现 */ }
```

**特点**:
- 统一接口，不同实现
- 运行时切换策略
- 支持链式调用

#### 1.3 工厂模式 (Factory Pattern)

**应用场景**: 服务器创建

**工厂函数**:
```go
func NewHttpServer(ip string, port int) *BeegoHttpServer
func NewHttpsServer(ip string, port int) *BeegoHttpsServer
```

**特点**:
- 封装创建逻辑
- 标准化接口
- 简化对象创建

#### 1.4 观察者模式 (Observer Pattern)

**应用场景**: HTTPS 证书更新

**实现**:
```go
type BeegoHttpsServer struct {
    restartChan chan CertInfo  // 事件通道
    isServerReady bool
}

// 观察者
go func() {
    for {
        select {
        case certInfo := <-restartChan:  // 监听事件
            // 处理证书更新事件
        }
    }
}()
```

**特点**:
- 事件驱动
- 异步处理
- 解耦事件发送和接收

#### 1.5 建造者模式 (Builder Pattern)

**应用场景**: 服务器配置

**实现**:
```go
server := https.NewHttpServer(ip, port)
server.InsertFilter("*", beego.BeforeRouter, filter)
server.Router("/video", &VideoController{})
server.Run()
```

**特点**:
- 链式调用
- 灵活配置
- 逐步构建

### 2. 架构原则

#### 2.1 单一职责原则 (SRP)

每个模块职责明确:
- **conf/**: 只负责配置管理
- **logger/**: 只负责日志记录
- **https/**: 只负责网络服务器
- **error/**: 只负责错误处理

#### 2.2 开闭原则 (OCP)

**对扩展开放，对修改关闭**:
- 通过接口定义，可以添加新的服务器实现
- 日志系统支持扩展新的日志处理器
- 配置系统可通过环境变量扩展

#### 2.3 里氏替换原则 (LSP)

**接口实现可替换**:
```go
var server BeegoServer

// 可以替换为 HTTP 或 HTTPS 实现
server = NewHttpServer(ip, port)
// 或
server = NewHttpsServer(ip, port)

server.Router("/test", &TestController{})
server.Run()
```

#### 2.4 接口隔离原则 (ISP)

**最小接口设计**:
- `BeegoServer` 接口只包含必要方法
- 认证接口 AuthGuard 只包含验证方法
- 存储接口 Storage 只包含存储操作

#### 2.5 依赖倒置原则 (DIP)

**依赖抽象而非具体**:
```go
// 依赖接口 BeegoServer
type BeegoHttpServer struct {
    server BeegoServer  // 依赖抽象
}

// 不依赖具体实现
```

### 3. 安全设计原则

#### 3.1 深度防御

多层安全措施:
1. 网络层: TLS 加密
2. 应用层: 证书验证
3. 审计层: 操作日志记录

#### 3.2 最小权限原则

权限严格控制:
- 证书验证拒绝不当的密钥用法
- 文件访问使用绝对路径
- 配置访问使用读写锁保护

#### 3.3 失败安全

默认安全配置:
- TLS 配置使用安全协议版本
- 拒绝弱加密算法
- 私钥加密存储

### 4. 性能优化原则

#### 4.1 并发安全

使用适当的同步机制:
- 配置读写使用 RWMutex
- 读操作并发执行
- 写操作独占访问

#### 4.2 资源管理

正确管理资源:
- 文件描述符及时关闭
- 证书通道缓存大小控制
- 避免资源泄漏

#### 4.3 缓存策略

合理的缓存机制:
- 配置单例模式
- 证书池复用
- IP 地址缓存

## 包间依赖关系

### 1. common 包内部依赖

```
https/
├── imports logger/ (日志记录)
└── imports error/ (错误处理)

logger/
├── imports constants/ (服务名称)
└── imports error/ (错误处理)

conf/
└── 无内部依赖

constants/
└── 无内部依赖

error/
└── 无内部依赖
```

### 2. common 包外部依赖

```
common/
├── imports github.com/beego/beego/v2/server/web (Beego 框架)
├── imports Go-chassis-extend/api/GSF/api (企业框架)
├── imports Go-chassis-extend/api/GSF/sdk/log/logapi (日志)
└── imports 标准库 (net, crypto/tls, sync, os, etc.)
```

### 3. 被 common 包依赖的模块

```
(使用 common 包的模块)
├── controllers/      - 使用 logger, https
├── routers/          - 使用 logger, https
├── services/         - 使用 logger, conf
├── remote/           - 使用 logger, error
├── storage/          - 使用 logger
├── tasks/            - use logger, conf
├── cert/             - use logger, https
└── main/             - use logger, conf, https, constants
```

### 4. 依赖图

```
                          ┌─────────────────────┐
                          │   External Libs     │
                          │                     │
                          │  - Beego            │
                          │  - Go-chassis       │
                          │  - Go stdlib        │
                          └──────────┬──────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────┐
│                        Package Dependencies                    │
│                                                               │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐              │
│  │          │    │ common/  │    │          │              │
│  │main      │◄───┤          ├───►│routers   │              │
│  │          │    │          │    │          │              │
│  └──────────┘    └──────────┘    └──────────┘              │
│                      │                                       │
│       ┌──────────────┼──────────────┐                       │
│       ▼              ▼              ▼                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                │
│  │controllers│ │ services │ │  remote  │                │
│  │          │  │          │  │          │                │
│  └──────────┘  └──────────┘  └──────────┘                │
│       │              │              │                       │
│       └──────────────┼──────────────┘                       │
│                      ▼                                       │
│              ┌──────────┐                                   │
│              │ storage  │                                   │
│              │ tasks    │                                   │
│              │ cert     │                                   │
│              └──────────┘                                   │
└──────────────────────────────────────────────────────────────┘
```

## 总结

MediaCacheService 的 common 包是一个设计良好、功能完整的基础设施层，具有以下特点：

### 优点：

1. **清晰的模块划分**: 各子包职责明确，易于理解和维护
2. **完善的日志系统**: 支持多级日志和审计日志
3. **强大的服务器支持**: HTTP/HTTPS 双协议，TLS 证书管理完善
4. **安全的设计**: 严格的证书验证，多层安全措施
5. **线程安全**: 配置管理使用读写锁，支持并发访问
6. **灵活的配置**: 支持环境变量，易于部署和调优

### 适用场景：

- 企业级微服务的基础设施建设
- 需要高安全性的网络服务
- 需要完善审计功能的应用
- 多网络环境部署
- 需要证书热更新的 HTTPS 服务

该 common 包为 MediaCacheService 提供了坚实的技术基础，是一个值得参考的 Go 项目基础设施实现典范。