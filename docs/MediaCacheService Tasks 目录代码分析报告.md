# MediaCacheService Tasks 目录代码分析报告

## 一、项目概述

MediaCacheService 是一个基于 Go 语言开发的媒体缓存服务，采用 Beego Web 框架和 ServiceComb 微服务架构，主要用于视频文件的本地缓存管理和远程服务调用。tasks 目录包含两个核心定时任务：磁盘扫描任务和数据清理任务，用于实现数据老化功能。

---

## 二、目录层级关系

```
D:\CloudCellular\MediaCacheService\
├── README.md                              # 项目说明文档
├── src\                                   # 源代码目录
│   ├── main.go                            # 应用程序主入口（270行）
│   ├── common\                            # 通用模块
│   │   ├── conf\config.go                 # 配置管理核心模块
│   │   ├── logger\                        # 日志模块
│   │   ├── error\                         # 错误处理
│   │   └── https\                         # HTTP/HTTPS服务器
│   ├── conf\                              # 配置文件目录
│   │   ├── app.conf                       # 应用基础配置
│   │   ├── chassis.yaml                   # ServiceComb框架配置
│   │   └── microservice.yaml              # 微服务属性配置
│   ├── controllers\                       # 控制器层
│   │   ├── controller.go                  # 基础控制器接口
│   │   ├── video_controller.go            # 视频控制器（165行）
│   │   └── filter.go                      # 过滤器
│   ├── service\                           # 服务层
│   │   ├── VideoService.go                # 视频服务核心实现（86行）
│   │   ├── AlarmService.go                # 告警服务实现（361行）
│   │   └── auth_service.go                # 认证服务实现
│   ├── storage\                           # 存储层
│   │   ├── storage.go                     # 存储接口定义
│   │   └── local_storage.go               # 本地存储实现（144行）
│   ├── tasks\                             # **任务模块（重点分析）**
│   │   ├── task_init.go                   # 任务初始化（41行）
│   │   ├── clearing_task.go               # 数据清理任务（63行）
│   │   └── scanning_task.go               # 数据扫描任务（43行）
│   ├── remote\                            # 远程服务调用
│   │   ├── remote.go                      # 远程服务接口（174行）
│   │   └── http.go                        # HTTP客户端（219行）
│   ├── routers\                           # 路由配置
│   │   └── router.go                      # HTTP路由配置
│   ├── util\                              # 工具模块
│   │   └── sys\sys.go                     # 系统操作工具
│   ├── cse\                               # ServiceComb服务
│   │   └── cse_helper.go                  # CSE助手
│   ├── cert\                              # 证书管理
│   │   └── cert.go                        # 证书实现
│   └── models\                            # 数据模型
│       └── resp\base.go                   # 响应类型定义
└── build\                                 # 构建相关
```

---

## 三、接口定义

### 1. controllers 包

#### IController 接口
```go
type IController interface {
    beego.ControllerInterface
    RouteInfo() RouteInfo
}
```
- **包路径**: `MediaCacheService/controllers`
- **说明**: 基础控制器接口，继承自 Beego 的 ControllerInterface，增加了路由信息方法
- **方法**:
  - `RouteInfo() RouteInfo`: 返回路由映射和过滤器配置

---

### 2. storage 包

#### Storage 接口
```go
type Storage interface {
    Cache(filePath string) (*FileInfo, error)
    Get(videoPath string) (io.ReadCloser, *FileInfo, error)
    Exist(filePath string) bool
}
```
- **包路径**: `MediaCacheService/storage`
- **说明**: 存储抽象接口，定义文件缓存的核心操作
- **方法**:
  - `Cache(filePath string) (*FileInfo, error)`: 缓存文件，返回文件信息
  - `Get(videoPath string) (io.ReadCloser, *FileInfo, error)`: 获取视频文件，返回读取流和文件信息
  - `Exist(filePath string) bool`: 检查文件是否存在

---

### 3. service 包

#### VideoService 接口
```go
type VideoService interface {
    GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
    Download(path string) (io.ReadCloser, int64, error)
}
```
- **包路径**: `MediaCacheService/service`
- **说明**: 视频服务接口，提供视频获取和下载功能
- **方法**:
  - `GetVideo(videoPath string)`: 获取视频流，优先从缓存读取
  - `Download(path string)`: 下载视频文件，返回流和文件大小

#### AuthService 接口
```go
type AuthService interface {
    ValidateIMEI(imei string, checkType string) (bool, error)
}
```
- **说明**: 认证服务接口，验证设备IMEI有效性
- **方法**:
  - `ValidateIMEI(imei string, checkType string)`: 验证IMEI和检查类型

#### AlarmService 接口
```go
type AlarmService interface {
    SendAlarm(alarmID, EventMessage string)
    ClearAlarm(alarmID, EventMessage string)
}
```
- **说明**: 告警服务接口，处理系统告警的发送和清除
- **方法**:
  - `SendAlarm(alarmID, EventMessage string)`: 发送告警
  - `ClearAlarm(alarmID, EventMessage string)`: 清除告警

---

### 4. remote 包

#### Remote 接口
```go
type Remote interface {
    GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
    PostValidateIMEI(IMEI string, CheckType string) (bool, error)
    GetGIDSAddress() (string, error)
}
```
- **包路径**: `MediaCacheService/remote`
- **说明**: 远程服务调用接口，对接外部 MUEN 服务
- **方法**:
  - `GetVideo(videoPath string)`: 从远程服务获取视频
  - `PostValidateIMEI(IMEI string, CheckType string)`: 验证IMEI
  - `GetGIDSAddress()`: 获取GIDS服务地址

#### HttpClient 接口
```go
type HttpClient interface {
    Get(url string, headers map[string]string) (*http.Response, error)
    GetWithRetry(url string, headers map[string]string, attempt int) (*http.Response, error)
    Post(url string, headers map[string]string, body interface{}) (*http.Response, error)
    PostWithRetry(url string, headers map[string]string, body interface{}, attempt int) (*http.Response, error)
}
```
- **说明**: HTTP客户端接口，支持重试机制
- **方法**:
  - `Get`: GET请求
  - `GetWithRetry`: 带重试的GET请求
  - `Post`: POST请求
  - `PostWithRetry`: 带重试的POST请求

---

### 5. util/sys 包

#### Interface 接口
```go
type Interface interface {
    SysDirSize(dirPath string) (int, error)
    DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
}
```
- **包路径**: `MediaCacheService/util/sys`
- **说明**: 系统操作接口，提供文件系统操作
- **方法**:
  - `SysDirSize(dirPath string)`: 获取目录大小（MB）
  - `DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration)`: 删除不活跃文件

---

### 6. cse 包

#### CSEHelper 接口
```go
type CSEHelper interface {
    GetServiceInstance(msKey base.MicroServiceKey) ([]base.MicroServiceInstance, error)
    ExtractIPPort(endpoint string) (string, error)
    GetAvailableEndpoints(msKey base.MicroServiceKey) (map[string]struct{}, error)
}
```
- **包路径**: `MediaCacheService/cse`
- **说明**: ServiceComb CSE助手接口，管理微服务实例和端点
- **方法**:
  - `GetServiceInstance`: 获取微服务实例
  - `ExtractIPPort`: 从端点提取IP和端口
  - `GetAvailableEndpoints`: 获取可用端点列表

---

## 四、结构体定义

### 1. controllers 包

#### RouteInfo
```go
type RouteInfo struct {
    RouteMapping map[string]string
    Filters      map[FilterAction]beego.FilterFunc
}
```
- **包路径**: `MediaCacheService/controllers`
- **说眀**: 路由信息配置结构体
- **字段**:
  - `RouteMapping map[string]string`: 路由映射表
  - `Filters map[FilterAction]beego.FilterFunc`: 过滤器函数映射

#### BaseController
```go
type BaseController struct {
    beego.Controller
}
```
- **嵌入**: `beego.Controller`
- **说明**: 基础控制器，继承自Beego控制器

#### VideoController
```go
type VideoController struct {
    BaseController
    videoService service.VideoService
    authService  service.AuthService
}
```
- **嵌入**: `BaseController`
- **字段**:
  - `videoService service.VideoService`: 视频服务实例
  - `authService service.AuthService`: 认证服务实例

---

### 2. common/conf 包

#### Config
```go
type Config struct {
    Logger      LoggerConfig      `flag:"log"`
    MediaCache  string            `flag:"cache" desc:"video local cache address"`
    HTTPTimeout int
    DataAging   DataAgingConfig   `flag:"data"`
}
```
- **标签**: `flag:"log"`, `flag:"cache"`, `flag:"data"`
- **字段**:
  - `Logger LoggerConfig`: 日志配置
  - `MediaCache string`: 视频缓存路径（默认 `/opt/mtuser/mcs/video`）
  - `HTTPTimeout int`: HTTP超时时间
  - `DataAging DataAgingConfig`: 数据老化配置

#### LoggerConfig
```go
type LoggerConfig struct {
    LogFile  string `flag:"file" desc:"log file path"`
    LogLevel string `flag:"level" desc:"log level: INFO/WARN/DEBUG"`
}
```
- **标签**: `flag:"file"`, `flag:"level"`
- **字段**:
  - `LogFile string`: 日志文件路径
  - `LogLevel string`: 日志级别（INFO/WARN/DEBUG）

#### DataAgingConfig
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
- **标签**: 各字段都有对应的flag标签
- **字段**:
  - `ScanningTaskPeriod string`: 扫描任务周期（默认1小时）
  - `ClearingTaskPeriod string`: 清理任务周期（默认24小时）
  - `ClearingTaskThreshold string`: 清理任务阈值（默认480GB）
  - `FileAccessInactiveThreshold string`: 文件访问不活跃阈值（天数）
  - `DeleteInactiveFileTimeout string`: 删除不活跃文件超时时间（秒）
  - `CacheAvailable bool`: 缓存可用状态标记

---

### 3. storage 包

#### FileInfo
```go
type FileInfo struct {
    Name             string
    Path             string
    Size             string
    ModifiedTime     time.Time
    Hash             string
    HasCached        bool
    ExtraWriteTarget io.WriteSeeker
    Finalizer        func()
}
```
- **说明**: 文件信息结构体，包含文件的元数据和状态
- **字段**:
  - `Name string`: 文件名
  - `Path string`: 文件路径
  - `Size string`: 文件大小
  - `ModifiedTime time.Time`: 修改时间
  - `Hash string`: 文件哈希值
  - `HasCached bool`: 是否已缓存
  - `ExtraWriteTarget io.WriteSeeker`: 额外的写入目标
  - `Finalizer func()`: 最终清理函数

#### localStorage
```go
type localStorage struct {
    name     string
    basePath string
}
```
- **说明**: 本地存储实现结构体
- **字段**:
  - `name string`: 存储名称
  - `basePath string`: 基础路径

---

### 4. service 包

#### VideoServiceImpl
```go
type VideoServiceImpl struct {
    remote  remote.Remote
    storage storage.Storage
    alarm   AlarmService
}
```
- **说明**: 视频服务实现
- **字段**:
  - `remote remote.Remote`: 远程服务客户端
  - `storage storage.Storage`: 存储接口
  - `alarm AlarmService`: 告警服务

#### AuthServiceImpl
```go
type AuthServiceImpl struct {
    remote remote.Remote
}
```
- **说明**: 认证服务实现
- **字段**:
  - `remote remote.Remote`: 远程服务客户端

#### AlarmEvent
```go
type AlarmEvent struct {
    AlarmID      string
    EventMessage string
    Type         base.GenerateOrClearType
}
```
- **说明**: 告警事件结构体
- **字段**:
  - `AlarmID string`: 告警ID
  - `EventMessage string`: 事件消息
  - `Type base.GenerateOrClearType`: 事件类型（生成或清除）

#### alarmServiceImpl
```go
type alarmServiceImpl struct {
    mu           sync.Mutex
    alarms       map[string]int64
    alarmManager base.CSPAlarmManager
}
```
- **说明**: 告警服务实现，实现AlarmService接口
- **字段**:
  - `mu sync.Mutex`: 互斥锁，保证线程安全
  - `alarms map[string]int64`: 告警记录映射
  - `alarmManager base.CSPAlarmManager`: 告警管理器

#### AlarmInfo
```go
type AlarmInfo struct {
    Location   string `json:"location,omitempty"`
    AppendInfo string `json:"appendInfo,omitempty"`
    AlarmId    string `json:"alarmId,omitempty"`
}
```
- **标签**: `json:"location,omitempty"`, `json:"appendInfo,omitempty"`, `json:"alarmId,omitempty"`
- **说明**: 告警信息结构体
- **字段**:
  - `Location string`: 位置信息
  - `AppendInfo string`: 附加信息
  - `AlarmId string`: 告警ID

#### AlarmResponse
```go
type AlarmResponse struct {
    Retdesc string      `json:"retdesc,omitempty"`
    Data    []AlarmInfo `json:"data,omitempty"`
    RetCode string      `json:"retcode,omitempty"`
}
```
- **标签**: 各字段都有json标签
- **说明**: 告警响应结构体
- **字段**:
  - `Retdesc string`: 响应描述
  - `Data []AlarmInfo`: 告警数据
  - `RetCode string`: 返回码

---

### 5. remote 包

#### HttpClientConfig
```go
type HttpClientConfig struct {
    Timeout time.Duration
}
```
- **说明**: HTTP客户端配置
- **字段**:
  - `Timeout time.Duration`: 超时时间

#### remoteImpl
```go
type remoteImpl struct {
    httpClient   HttpClient
    cseHelper    cse.CSEHelper
    configCenter ConfigCenterService
}
```
- **说明**: 远程服务实现，实现Remote接口
- **字段**:
  - `httpClient HttpClient`: HTTP客户端
  - `cseHelper cse.CSEHelper`: CSE助手
  - `configCenter ConfigCenterService`: 配置中心服务

#### httpClientImpl
```go
type httpClientImpl struct {
    client *http.Client
}
```
- **说明**: HTTP客户端实现
- **字段**:
  - `client *http.Client`: Go标准HTTP客户端

---

### 6. util/sys 包

#### sysImpl
```go
type sysImpl struct {}
```
- **说明**: 系统操作实现，实现Interface接口
- **字段**: 无

---

### 7. main.go

#### GracefulExitHandler
```go
type GracefulExitHandler struct {}
```
- **说明**: 优雅退出处理器，实现gseapi.ExitHandler接口
- **字段**: 无

---

## 五、函数实现详解

### 1. tasks 包

#### InitCronTasks()
- **文件**: `src/tasks/task_init.go`
- **行数**: 第22-40行
- **函数签名**:
  ```go
  func InitCronTasks()
  ```
- **参数**: 无
- **返回值**: 无
- **核心逻辑实现**:
  1. 获取配置实例: `cf := conf.Instance()`
  2. 解析扫描任务周期:
     - 从配置读取 `cf.DataAging.ScanningTaskPeriod`
     - 转换为整数，失败时使用默认值1小时
     - 记录错误日志
  3. 解析清理任务周期:
     - 从配置读取 `cf.DataAging.ClearingTaskPeriod`
     - 转换为整数，失败时使用默认值24小时
     - 记录错误日志
  4. 创建定时任务:
     - 扫描任务: `task.AddTask("scanning_task", task.NewTask("scanning_task", fmt.Sprintf("0 0 */%d * * *", stp), scanningTask))`
     - 清理任务: `task.AddTask("clearing_task", task.NewTask("clearing_task", fmt.Sprintf("0 0 */%d * * *", ctp), clearingTask))`
     - Cron表达式格式: `0 0 */N * * *` 表示每N小时执行一次
  5. 程序启动时立即执行清理任务:
     - 调用 `clearingTask(goctx.TODO())`
     - 执行失败时记录错误日志
- **调用关系**: 被 `main.go` 中的主函数调用
- **副作用**: 设置定时任务，修改全局任务注册表
- **错误处理**:
  - 配置解析失败时使用默认值并记录日志
  - 启动时清理任务执行失败时记录错误日志

#### clearingTask(ctx goctx.Context) error
- **文件**: `src/tasks/clearing_task.go`
- **行数**: 第16-62行
- **函数签名**:
  ```go
  func clearingTask(ctx goctx.Context) error
  ```
- **参数**:
  - `ctx goctx.Context`: 上下文对象，用于任务取消和超时控制
- **返回值**: `error` - 操作错误或nil表示成功
- **核心逻辑实现**:
  1. **上下文取消检查**:
     ```go
     select {
     case <-ctx.Done():
         logger.Infof("tasks is stop")
         return nil
     default:
     }
     ```
     - 检测上下文是否已取消，如果取消则返回nil
  2. **获取配置**:
     ```go
     cf := conf.Instance()
     cachePath := cf.MediaCache
     ```
     - 获取全局配置实例和缓存路径
  3. **解析清理阈值**:
     ```go
     cleanOriThreshold, err := strconv.Atoi(cf.DataAging.ClearingTaskThreshold)
     cleanThreshold := int(math.Floor(float64(cleanOriThreshold) * 0.75))
     ```
     - 从配置读取清理阈值（默认480GB）
     - 计算清理触发阈值（原阈值的75%，即360GB）
  4. **获取当前目录大小**:
     ```go
     sysI := sys.NewFunc()
     usedMb, err := sysI.SysDirSize(cachePath)
     ```
     - 创建系统操作工具实例
     - 获取缓存目录当前使用大小（MB）
  5. **判断是否需要清理**:
     ```go
     if usedMb <= cleanThreshold {
         logger.Infof("usedMb %d <= cleanThreshold %d, skip cleaning task")
         return nil
     }
     ```
     - 如果使用空间 <= 清理阈值，跳过清理
  6. **解析文件不活跃阈值**:
     ```go
     threshold, err := strconv.Atoi(cf.DataAging.FileAccessInactiveThreshold)
     ```
     - 解析文件访问不活跃阈值（天数）
  7. **解析删除操作超时时间**:
     ```go
     timeout, err := strconv.Atoi(cf.DataAging.DeleteInactiveFileTimeout)
     ```
     - 解析删除操作超时时间（秒）
  8. **执行文件删除操作**:
     ```go
     if err = sysI.DeleteInactiveFile(cachePath, threshold, time.Duration(timeout)*time.Second); err != nil {
         logger.Errorf("failed to exec DeleteInactiveFile: %v", err)
         return err
     }
     ```
     - 调用系统工具删除不活跃文件
     - 失败时记录错误并返回
  9. **记录成功日志**:
     ```go
     logger.Infof("success delete inactive file, threshold is %d", threshold)
     ```
  10. **链式调用扫描任务更新状态**:
      ```go
      return scanningTask(ctx)
      ```
      - 清理完成后主动调用扫描任务更新缓存可用状态
- **调用关系**:
  - 被 `InitCronTasks()` 注册为定时任务
  - 清理完成后主动调用 `scanningTask()`
  - 内部调用 `sys.SysDirSize()` 和 `sys.DeleteInactiveFile()`
- **副作用**:
  - 删除文件系统中的不活跃文件
  - 修改磁盘空间使用情况
  - 更新缓存可用状态（通过链式调用扫描任务）
- **配置依赖**:
  - `MediaCache`: 缓存目录路径
  - `ClearingTaskThreshold`: 磁盘使用阈值（MB）
  - `FileAccessInactiveThreshold`: 文件不活跃阈值（天数）
  - `DeleteInactiveFileTimeout`: 删除操作超时时间（秒）

#### scanningTask(ctx goctl.Context) error
- **文件**: `src/tasks/scanning_task.go`
- **行数**: 第15-42行
- **函数签名**:
  ```go
  func scanningTask(ctx goctx.Context) error
  ```
- **参数**:
  - `ctx goctx.Context`: 上下文对象，用于任务取消和超时控制
- **返回值**: `error` - 操作错误或nil表示成功
- **核心逻辑实现**:
  1. **上下文取消检查**:
     ```go
     select {
     case <-ctx.Done():
         logger.Infof("tasks is stop")
         return nil
     default:
     }
     ```
     - 检测上下文是否已取消，如果取消则返回nil
  2. **获取配置**:
     ```go
     cf := conf.Instance()
     cachePath := cf.MediaCache
     ```
     - 获取全局配置实例和缓存路径
  3. **解析扫描阈值**:
     ```go
     oriThreshold, err := strconv.Atoi(cf.DataAging.ClearingTaskThreshold)
     threshold := int(math.Floor(float64(oriThreshold) * 0.85))
     ```
     - 从配置读取阈值（默认480GB）
     - 计算可用状态判断阈值（原阈值的85%，即408GB）
  4. **获取当前目录大小**:
     ```go
     sysI := sys.NewFunc()
     usedMb, err := sysI.SysDirSize(cachePath)
     ```
     - 创建系统操作工具实例
     - 获取缓存目录当前使用大小（MB）
  5. **判断缓存可用状态**:
     ```go
     cacheAvailable := usedMb <= threshold
     ```
     - 如果使用空间 <= 85%阈值，缓存可用
     - 否则缓存不可用
  6. **更新可用状态**:
     ```go
     conf.SetCacheAvailable(cacheAvailable)
     ```
     - 将缓存可用状态更新到全局配置
  7. **记录日志**:
     ```go
     logger.Infof("success scanning dir size, current cache available stat is :%v, used is %d, threshold is %d", cacheAvailable, usedMb, threshold)
     ```
     - 输出当前状态、使用空间和阈值
- **调用关系**:
  - 被 `InitCronTasks()` 注册为定时任务
  - 被 `clearingTask()` 链式调用（清理完成后）
  - 内部调用 `sys.SysDirSize()` 和 `conf.SetCacheAvailable()`
- **副作用**:
  - 更新全局配置中的缓存可用状态
  - 影响后续视频服务的缓存使用决策
- **配置依赖**:
  - `MediaCache`: 缓存目录路径
  - `ClearingTaskThreshold`: 磁盘使用阈值（MB）

---

### 2. util/sys 包

#### SysDirSize(dirPath string) (int, error)
- **文件**: `src/util/sys/sys.go`
- **函数签名**:
  ```go
  func SysDirSize(dirPath string) (int, error)
  ```
- **参数**:
  - `dirPath string`: 要测量的目录路径
- **返回值**: `(int, error)` - 目录大小（MB）或错误
- **核心逻辑实现**:
  1. 执行系统命令获取目录大小:
     ```go
     cmd := exec.Command("du", "-sm", dirPath)
     ```
     - 使用 `du -sm` 命令获取目录大小（MB）
  2. 执行命令并获取输出:
     ```go
     output, err := cmd.CombinedOutput()
     ```
  3. 解析命令输出:
     ```go
     parts := strings.Fields(string(output))
     if len(parts) < 1 {
         return 0, fmt.Errorf("invalid output from du command")
     }
     sizeStr := parts[0]
     ```
  4. 转换为整数并返回:
     ```go
     size, err := strconv.Atoi(sizeStr)
     if err != nil {
         return 0, err
     }
     return size, nil
     ```
- **调用关系**: 被 `clearingTask()` 和 `scanningTask()` 调用
- **副作用**: 无（只读操作）
- **依赖**: Unix/Linux系统 `du` 命令

#### DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
- **文件**: `src/util/sys/sys.go`
- **函数签名**:
  ```go
  func DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
  ```
- **参数**:
  - `dirPath string`: 要搜索的目录路径
  - `threshold int`: 文件不活跃阈值（天数）
  - `timeout time.Duration`: 操作超时时间
- **返回值**: `error` - 操作错误或nil表示成功
- **核心逻辑实现**:
  1. 创建带超时的上下文:
     ```go
     ctx, cancel := context.WithTimeout(context.Background(), timeout)
     defer cancel()
     ```
  2. 构造find命令查找不活跃文件:
     ```go
     cmd := exec.Command("bash", "-c",
         fmt.Sprintf("find %s -type f -atime +%d -exec rm -f {} \\;", dirPath, threshold))
     ```
     - 使用 `find` 命令查找访问时间超过阈值天数的文件
     - `-type f`: 只查找文件
     - `-atime +threshold`: 访问时间超过阈值
     - `-exec rm -f {} \;`: 删除找到的文件
  3. 设置命令上下文:
     ```go
     cmd = exec.CommandContext(ctx, "bash", "-c",
         fmt.Sprintf("find %s -type f -atime +%d -exec rm -f {} \\;", dirPath, threshold))
     ```
  4. 执行命令:
     ```go
     output, err := cmd.CombinedOutput()
     if err != nil {
         return fmt.Errorf("failed to delete inactive files: %v, output: %s", err, output)
     }
     ```
  5. 返回成功:
     ```go
     return nil
     ```
- **调用关系**: 被 `clearingTask()` 调用
- **副作用**: 永久删除文件系统中的文件（破坏性操作）
- **依赖**: Unix/Linux系统 `find` 命令

---

### 3. common/conf 包

#### Instance() *Config
- **文件**: `src/common/conf/config.go`
- **函数签名**:
  ```go
  func Instance() *Config
  ```
- **参数**: 无
- **返回值**: `*Config` - 配置实例指针
- **核心逻辑实现**:
  1. 使用单例模式返回配置实例
  2. 如果实例不存在则创建新实例
  3. 从配置文件加载配置数据
- **调用关系**: 被所有需要访问配置的函数调用
- **副作用**: 无（初始化时创建全局实例）

#### SetCacheAvailable(available bool)
- **文件**: `src/common/conf/config.go`
- **函数签名**:
  ```go
  func SetCacheAvailable(available bool)
  ```
- **参数**:
  - `available bool`: 缓存可用状态
- **返回值**: 无
- **核心逻辑实现**:
  1. 获取全局配置实例
  2. 使用写锁保证线程安全:
     ```go
     cf.mu.Lock()
     defer cf.mu.Unlock()
     cf.DataAging.CacheAvailable = available
     ```
- **调用关系**: 被 `scanningTask()` 调用
- **副作用**: 修改全局配置状态，影响后续缓存使用决策

---

## 六、函数调用关系图

### 1. 整体调用关系

```
main()
    ↓
InitCronTasks()
    ↓
    ├─→ 创建扫描任务定时器 (scanning_task, 每小时)
    │       ↓
    │   scanningTask()
    │       ↓
    │   ├─→ sys.SysDirSize() (获取目录大小)
    │   └─→ conf.SetCacheAvailable() (更新可用状态)
    │
    └─→ 创建清理任务定时器 + 立即执行 (clearing_task, 每天或启动时)
            ↓
        clearingTask()
            ↓
        ├─→ sys.SysDirSize() (获取目录大小)
        ├─→ sys.DeleteInactiveFile() (删除不活跃文件)
        │       ↓
        │   执行 find -atime 命令删除文件
        └─→ scanningTask() (链式调用更新状态)
                ↓
            conf.SetCacheAvailable()
```

### 2. 详细调用链路

#### 数据老化流程调用链:
```
clearingTask()
    ├── conf.Instance() → 获取配置
    ├── sys.SysDirSize() → 获取当前目录大小
    │       └── exec.Command("du -sm") → 执行系统命令
    ├── 判断 usedMb <= cleanThreshold
    │       └── true: 跳过清理，直接返回
    │       └── false: 继续清理
    ├── sys.DeleteInactiveFile() → 删除不活跃文件
    │       └── exec.Command("find -atime -exec rm") → 执行删除命令
    └── scanningTask() → 链式调用更新状态
            ├── sys.SysDirSize() → 再次获取目录大小
            └── conf.SetCacheAvailable() → 更新可用状态
```

#### 磁盘扫描流程调用链:
```
scanningTask()
    ├── conf.Instance() → 获取配置
    ├── sys.SysDirSize() → 获取当前目录大小
    │       └── exec.Command("du -sm") → 执行系统命令
    ├── 判断 usedMb <= threshold (85%)
    │       └── true: 缓存可用
    │       └── false: 缓存不可用
    └── conf.SetCacheAvailable() → 更新状态
```

### 3. 配置更新流程:
```
conf.SetCacheAvailable(available bool)
    ├── 获取配置实例 cf.Instance()
    │       ↓
    ├── 获取写锁 cf.mu.Lock()
    │       ↓
    ├── 更新状态 cf.DataAging.CacheAvailable = available
    │       ↓
    └── 释放锁 cf.mu.Unlock()
```

---

## 七、主要业务流程执行路径

### 1. 程序启动流程:

```
main()
    ↓
1. 初始化GSF框架
    ↓
2. 加载配置文件
    │   ├── app.conf (应用配置)
    │   ├── chassis.yaml (ServiceComb框架配置)
    │   └── microservice.yaml (微服务属性配置)
    ↓
3. tasks.InitCronTasks()
    │   ├── 解析 ScanningTaskPeriod (默认1小时)
    │   ├── 解析 ClearingTaskPeriod (默认24小时)
    │   ├── 注册扫描任务定时器
    │   ├── 注册清理任务定时器
    │   └── 立即执行清理任务 clearingTask()
    │           ↓
    │       1. 获取当前目录大小
    │       2. 判断是否需要清理
    │       3. 如需要，删除不活跃文件
    │       4. 扫描任务扫描目录大小，更新缓存可用状态
    ↓
4. 初始化HTTP/HTTPS服务器
    ├── 内部HTTP服务器 (端口9996/9997)
    └── 外部HTTP服务器 (端口9990/9991)
    ↓
5. 注册路由
    └── router.go 配置路由映射
    ↓
6. 注册微服务
    └── ServiceComb服务注册
    ↓
7. 启动服务并监听
```

### 2. 任务执行流程:

#### 扫描任务 (每小时执行):
```
scanningTask()
    ↓
1. 检查上下文是否取消
    ↓
2. 获取配置和缓存路径
    ↓
3. 计算扫描阈值 (原阈值的85%)
    ↓
4. 获取当前目录大小 (sys.SysDirSize)
    ↓
5. 判断: usedMb <= threshold?
    ├─→ true: 缓存可用
    └─→ false: 缓存不可用
    ↓
6. 更新全局配置状态 (conf.SetCacheAvailable)
    ↓
7. 记录日志
```

#### 清理任务 (每天执行或启动时执行):
```
clearingTask()
    ↓
1. 检查上下文是否取消
    ↓
2. 获取配置和缓存路径
    ↓
3. 计算清理触发阈值 (原阈值的75%)
    ↓
4. 获取当前目录大小 (sys.SysDirSize)
    ↓
5. 判断: usedMb <= cleanThreshold?
    ├─→ true: 跳过清理，直接返回
    └─→ false: 继续执行清理
        ↓
        1. 解析文件不活跃阈值
        2. 解析删除操作超时时间
        3. 删除不活跃文件 (sys.DeleteInactiveFile)
        4. 记录成功日志
    ↓
6. 链式调用扫描任务 (scanningTask)
    ↓
7. 更新缓存可用状态并返回
```

### 3. 视频请求处理流程:
```
HTTP请求 (GET /video)
    ↓
VideoController.GetVideo()
    ↓
VideoService.GetVideo()
    ↓
1. 检查缓存可用状态 (conf.DataAging.CacheAvailable)
    ├─→ false: 拒绝请求，返回错误
    └─→ true: 继续处理
        ↓
        1. 检查本地缓存 (storage.Exist)
        │   ├─→ 存在: 直接返回本地文件
        │   └─→ 不存在: 从远程服务获取
        │       ↓
        │   2. 调用远程服务 (remote.GetVideo)
        │       │   ↓
        │       │   1. 通过CSE服务发现获取MUEN服务地址
        │       │   2. 发起HTTP请求下载视频
        │       │   3. 返回视频流
        │       ↓
        │   3. 缓存到本地 (storage.Cache)
        │       ↓
        │   4. 返回视频流
        ↓
2. 返回文件信息 (FileInfo)
```

---

## 八、关键函数副作用和状态变更

### 1. clearingTask() - 清理任务
**主要副作用**:
- **文件系统变更**: 永久删除访问时间超过阈值的文件（破坏性操作）
- **磁盘空间减少**: 删除文件后会释放磁盘空间
- **配置状态更新**: 通过链式调用扫描任务更新缓存可用状态
- **日志记录**: 记录清理操作的结果和相关信息

**状态变更**:
- 修改文件系统内容（文件被删除）
- 降低缓存目录的磁盘使用量
- 间接影响 `conf.DataAging.CacheAvailable` 状态
- 可能触发告警（如果删除失败或磁盘空间不足）

**配置依赖**:
```
清除触发阈值 = ClearingTaskThreshold × 0.75 (例如: 480GB × 0.75 = 360GB)
文件不活跃阈值 = FileAccessInactiveThreshold (例如: 7天)
删除操作超时 = DeleteInactiveFileTimeout (例如: 300秒)
```

### 2. scanningTask() - 扫描任务
**主要副作用**:
- **配置状态更新**: 修改全局配置中的缓存可用状态
- **日志记录**: 记录当前磁盘使用状态

**状态变更**:
- 更新 `conf.DataAging.CacheAvailable` 布尔值
  - `true`: 当 usedMb ≤ threshold (85%)
  - `false`: 当 usedMb > threshold (85%)
- 影响后续视频服务的缓存使用决策

**配置依赖**:
```
可用状态阈值 = ClearingTaskThreshold × 0.85 (例如: 480GB × 0.85 = 408GB)
```

### 3. SysDirSize() - 目录大小计算
**副作用**: 无（纯只读操作）
**状态变更**: 无
**依赖**: Unix/Linux系统 `du -sm` 命令

### 4. DeleteInactiveFile() - 删除不活跃文件
**主要副作用**:
- **文件系统变更**: 永久删除符合条件的文件（破坏性操作）
- **磁盘空间减少**: 可能释放大量磁盘空间
- **操作超时**: 可能因为超时被取消

**状态变更**:
- 删除访问时间超过阈值天数的所有文件
- 没有直接修改配置状态
- 通过间接方式影响磁盘使用量

**执行的超时控制**:
- 使用 context.WithTimeout 实现超时控制
- 超时后自动取消删除操作

### 5. 配置更新函数
#### conf.SetCacheAvailable(available bool)
**副作用**:
- **状态更新**: 修改全局配置中的缓存可用状态
- **线程同步**: 使用读写锁保证并发访问时的数据一致性

**状态变更**:
- 直接设置 `cf.DataAging.CacheAvailable = available`
- 状态变更立即可见，但需要通过读锁访问

---