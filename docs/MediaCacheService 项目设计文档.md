# MediaCacheService 项目设计文档

## 1. 系统架构设计

### 1.1 架构概述

MediaCacheService 是一个基于 Go 语言开发的微服务，采用分层架构设计，提供媒体内容缓存和流式传输功能。系统主要服务于视频媒体内容的高效缓存和分发。

### 1.2 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      客户端请求                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   负载均衡/限流层                             │
│              (Greatwall 过载控制)                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   控制器层                                   │
│    ┌──────────────────────────────────────────────┐         │
│    │         VideoController                     │         │
│    │  - /video/* : 视频流式传输                   │         │
│    │  - /download/* : 文件下载                   │         │
│    │  - /test : 健康检查                         │         │
│    └──────────────────────────────────────────────┘         │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │                           │
        ▼                           ▼
┌───────────────┐           ┌───────────────┐
│  AuthService  │           │ VideoService  │
│   IMEI鉴权    │           │  视频缓存逻辑 │
└───────────────┘           └───────┬───────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
           ┌───────────────┐               ┌───────────────┐
           │   Remote      │               │   Storage     │
           │  远程服务集成  │               │   本地缓存     │
           └───────┬───────┘               └───────┬───────┘
                   │                               │
         ┌─────────┴─────────┐           ┌─────────┴─────────┐
         │                   │           │                   │
         ▼                   ▼           │                   │
    ┌─────────┐         ┌─────────┐     │                   │
    │  MUEN   │         │  GIDS   │     │                   │
    │ 媒体服务器│         │ 认证服务 │     │                   │
    └─────────┘         └─────────┘     │                   │
                                        │                   │
                                        ▼                   │
                               ┌───────────────┐           │
                               │  文件系统存储   │           │
                               └───────────────┘           │
                                                          │
                                                          ▼
                                                 ┌───────────────┐
                                                 │  告警服务      │
                                                 │  定时任务      │
                                                 └───────────────┘
```

### 1.3 核心设计原则

1. **分层清晰**：控制器、服务层、存储层职责明确
2. **接口抽象**：各层通过接口定义，便于扩展和替换实现
3. **缓存优先**：优先命中本地缓存，缓存未命中时回源获取
4. **服务治理**：集成华为CSP微服务框架，提供服务发现、配置管理、监控告警
5. **高可用**：支持多端口监听（内部/外部）、优雅关闭、故障告警

### 1.4 技术栈

| 技术组件 | 版本/类型 | 用途 |
|---------|----------|------|
| Go语言 | 1.20+ | 主开发语言 |
| Beego v2 | v2.1.0 | v0.0.0 | 云服务框架/微服务治理 |
| CSP SDK | - | 华为云服务集成 |
| Greatwall SDK | v1.9.6 | 过载定义 | 远程服务调用 |

---

## 2. 目录结构设计

### 2.1 完整目录树

```
CacheMediaService/
├── src/                                # 源代码目录
│   ├── main.go                         # 程序入口
│   ├── go.mod                          # Go模块定义
│   ├── go.sum                          # 依赖锁定文件
│   │
│   ├── cert/                           # 证书管理
│   │   ├── cert.go                     # 证书处理逻辑
│   │   └── cert_test.go                # 证书单元测试
│   │
│   ├── common/                         # 公共组件
│   │   ├── conf/                       # 配置管理
│   │   │   └── config.go               # 配置结构定义和加载
│   │   ├── constants/                  # 常量定义
│   │   │   ├── base.go                 # 基础常量
│   │   │   └── retcode/                # 返回码常量
│   │   │       └── retcode.go          # 错误码定义
│   │   ├── error/                      # 错误处理
│   │   │   └── error.go                # 错误定义
│   │   ├── https/                      # HTTP服务
│   │   │   ├── https_server.go         # HTTPS服务实现
│   │   │   ├── http_server.go          # HTTP服务实现
│   │   │   ├── tls.go                  # TLS配置
│   │   │   └── *_test.go               # 相关测试
│   │   └── logger/                     # 日志组件
│   │       ├── logger.go               # 日志接口和实现
│   │       └── auditlog.go             # 审计日志
│   │
│   ├── controllers/                    # 控制器层
│   │   ├── controller.go               # 基础控制器
│   │   ├── video_controller.go         # 视频相关控制器
│   │   ├── filter.go                   # 过滤器(过载防护)
│   │   └── *_test.go                   # 控制器测试
│   │
│   ├── cse/                            # CSE服务集成
│   │   └── cse_helper.go               # CSE辅助功能
│   │
│   ├── models/                         # 数据模型
│   │   └ resp──/                       # 响应模型
│   │       └── base.go                 # 基础响应结构
│   │
│   ├── remote/                         # 远程服务集成
│   │   ├── remote.go                   # 远程服务接口和实现
│   │   ├── client.go                   # HTTP客户端
│   │   ├── config_center.go            # 配置中心集成
│   │   └── http.go                     # HTTP工具
│   │
│   ├── routers/                        # 路由注册
│   │   ├── router.go                   # 路由定义
│   │   └── beego_router.go             # Beego路由适配
│   │
│   ├── service/                        # 业务服务层
│   │   ├── VideoService.go             # 视频服务
│   │   ├── auth_service.go             # 认证服务
│   │   └── AlarmService.go             # 告警服务
│   │
│   ├── storage/                        # 存储层
│   │   ├── storage.go                  # 存储接口定义
│   │   └── local_storage.go            # 本地文件存储实现
│   │
│   ├── tasks/                          # 定时任务
│   │   ├── task_init.go                # 任务初始化
│   │   ├── scanning_task.go            # 扫描任务
│   │   ├── clearing_task.go            # 清理任务
│   │   └── *_test.go                   # 任务测试
│   │
│   └── util/                           # 工具类
│       ├── flag/                       # 命令行参数解析
│       │   └── flags.go
│       ├── sys/                        # 系统工具
│       │   └── sys.go
│       └── response_util.go            # 响应工具
│
├── build/                              # 构建相关
├── docs/                               # 文档目录
└── README.md                           # 项目说明
```

### 2.2 目录职责说明

| 目录 | 职责 | 关键文件 |
|------|------|----------|
| `main.go` | 程序入口，初始化框架，启动服务 | 应用启动逻辑 |
| `common/` | 公共基础设施组件 | 配置、日志、HTTP服务等 |
| `controllers/` | HTTP请求处理和路由映射 | 每个API端点的实现 |
| `service/` | 业务逻辑层，协调多个组件完成业务功能 | VideoService, AuthService等 |
| `storage/` | 数据存储抽象和实现 | 本地文件缓存管理 |
| `remote/` | 外部服务集成 | MUEN、GIDS等服务调用 |
| `models/` | 数据结构定义 | API响应、实体定义 |
| `routers/` | 路由注册和配置 | API路径映射 |
| `tasks/` | 后台定时任务 | 缓存清理、监控扫描 |
| `util/` | 工具函数和辅助类 | 系统操作、响应格式化等 |
| `cert/` | 证书管理 | TLS证书加载和更新 |

---

## 3. 核心实体设计

### 3.1 控制器实体

#### BaseController
```go
type BaseController struct {
    beego.Controller
}

// 核心方法保持不变
- QueryParameter(name string) string           // 获取查询参数
- PathParameter(name string) string            // 获取路径参数  
- WriteHeaderAndJSON(status int, v interface{}, contentType string) error
- AddHeader(header, value string)              // 添加响应头
- ResponseWriter() http.ResponseWriter         // 获取响应写入器
- Request() *http.Request                      // 获取请求对象
- DownloadFile(file string)                    // 文件下载
- OK(data interface{})                         // 成功响应
- Failed(data resp.BaseResponse)               // 失败响应
- InternalServiceError()                       // 内部错误响应
```

#### VideoController
```go
type VideoController struct {
    BaseController
    videoService service.VideoService
    authService  service.AuthService
}

// 路由信息保持不变
RouteMapping: map[string]string{
    "/video/*":    "GET:GetVideo",      // 视频流式传输
    "/download/*": "GET:Download",      // 文件下载
    "/test":       "GET:Test",          // 健康检查
}

// 核心方法
- Prepare()                                   // 初始化服务实例
- GetVideo()                                   // 处理视频获取请求
- Download()                                   // 处理文件下载请求
- Test()                                       // 健康检查接口
- FormatHTTPHeader(fileInfo *storage.FileInfo) // 设置HTTP响应头
- addCorsHeader()                              // 添加CORS头
```

### 3.2 服务接口实体

#### VideoService
```go
type VideoService interface {
    GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
    Download(path string) (io.ReadCloser, int64, error)
}

type VideoServiceImpl struct {
    remote  remote.Remote
    storage storage.Storage
    alarm   AlarmService
}
```

#### AuthService
```go
type AuthService interface {
    ValidateIMEI(imei string, checkType string) (bool, error)
}

type AuthServiceImpl struct {
    remote remote.Remote
}
```

#### AlarmService
```go
type AlarmService interface {
    SendAlarm(alarmID, EventMessage string)
    ClearAlarm(alarmID, EventMessage string)
}

type alarmServiceImpl struct {
    mu           sync.Mutex
    alarms       map[string]int64
    alarmManager base.CSPAlarmManager
}
```

### 3.3 存储接口实体

#### Storage Interface
```go
type Storage interface {
    Cache(filePath string) (*FileInfo, error)    // 缓存文件
    Get(videoPath string) (io.ReadCloser, *FileInfo, error)  // 获取文件
    Exist(filePath string) bool                   // 检查文件是否存在
}
```

#### FileInfo 实体（保持不变）
```go
type FileInfo struct {
    Name             string              // 文件名
    Path             string              // 文件路径
    Size             string              // 文件大小(字符串格式)
    ModifiedTime     time.Time           // 修改时间
    Hash             string              // MD5哈希值
    HasCached        bool                // 是否已缓存
    ExtraWriteTarget io.WriteSeeker      // 额外写入目标
    Finalizer        func()              // 清理函数
}

// 核心方法
- AddFinalizer(newFinalizer func())  // 添加清理回调
```

### 3.4 远程服务实体

#### Remote Interface
```go
type Remote interface {
    GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
    PostValidateIMEI(IMEI string, CheckType string) (bool, error)
    GetGIDSAddress() (string, error)
}

type remoteImpl struct {
    httpClient   HttpClient
    cseHelper    cse.CSEHelper
    configCenter ConfigCenterService
}
```

### 3.5 配置实体

#### Config 结构（保持不变）
```go
type Config struct {
    Logger      LoggerConfig      `flag:"log"`
    MediaCache  string            `flag:"cache" desc:"video local cache address"`
    HTTPTimeout int               // HTTP超时时间(秒)
    DataAging   DataAgingConfig   `flag:"data"`
}

type LoggerConfig struct {
    LogFile  string `flag:"file" desc:"log file path"`
    LogLevel string `flag:"level" desc:"log level: INFO/WARN/DEBUG"`
}

type DataAgingConfig struct {
    ScanningTaskPeriod          string `flag:"scanning_task_period"`           // 扫描任务周期(小时)
    ClearingTaskPeriod          string `flag:"clearing_task_period"`           // 清理任务周期(小时)  
    ClearingTaskThreshold       string `flag:"clearing_task_threshold"`        // 触发清理的磁盘阈值(MB)
    FileAccessInactiveThreshold string `flag:"file_access_inactive_threshold"` // 文件未活跃阈值(天)
    DeleteInactiveFileTimeout   string `flag:"delete_inactive_file_timeout"`   // 删除超时(秒)
    CacheAvailable              bool                                                // 缓存可用状态
}
```

### 3.6 响应模型实体（保持完全一致）

#### BaseResponse
```go
type BaseResponse struct {
    Code    int    `json:"code"`    // 响应码
    Message string `json:"msg"`     // 响应消息
}
```

#### DataResponse
```go
type DataResponse struct {
    BaseResponse
    Data interface{} `json:"data"`  // 响应数据
}
```

### 3.7 路由信息实体

#### RouteInfo
```go
type RouteInfo struct {
    RouteMapping map[string]string                // 路由映射 {路径: "方法:处理函数"}
    Filters      map[FilterAction]beego.FilterFunc // 过滤器映射
}

type FilterAction string

const (
    Before FilterAction = "before"  // 前置过滤器
    After  FilterAction = "after"   // 后置过滤器
)
```

---

## 4. API接口设计

### 4.1 API清单与兼容性保证

#### 完整API路由表（必须保持现有接口签名和路径）

| HTTP方法 | 路径 | 处理函数 | 功能描述 | 请求参数 | 响应格式 |
|---------|------|----------|----------|----------|----------|
| GET | `/video/*` | GetVideo | 视频流式传输 | 查询参数: `imei`, `checkType`; 路径参数: `:splat` | 视频流(二进制) + 特定HTTP头 |
| GET | `/download/*` | Download | 文件下载 | 路径参数: `:splat` | 文件流(二进制) + DataResponse |
| GET | `/test` | Test | 健康检查 | 无 | DataResponse |

#### 中间件/过滤器（必须保持现有行为）

| 过滤器 | 应用范围 | 功能 | 配置键 |
|--------|----------|------|--------|
| OverLoadFilter | `/*` (全局) | 过载防护，当流量过大时返回429 | APIService |

### 4.2 详细接口定义

#### 4.2.1 视频流式传输接口

**端点:** `GET /video/*`

**功能说明:** 从缓存或MUEN服务器获取视频内容并流式传输给客户端，支持HLS(m3u8/ts)和MP4格式。

**请求参数:**

| 参数名 | 位置 | 类型 | 必填 | 描述 |
|--------|------|------|------|------|
| `imei` | Query参数 | string | 是 | 设备IMEI号，用于鉴权 |
| `checkType` | Query参数 | string | 否 | 鉴权类型 |
| `:splat` | 路径参数 | string | 是 | 视频文件相对路径，如 `/media/2025/video.mp4` |

**处理流程:**
1. 参数解析和安全校验
2. 调用AuthService验证IMEI有效性
3. 验证失败则返回错误，继续则执行下一步
4. 调用VideoService获取视频内容(含本地缓存逻辑)  
5. 设置合适的Content-Type和缓存控制HTTP头
6. 添加CORS支持头
7. 流式传输视频内容到客户端

**成功响应:**
- HTTP状态码: `200 OK`
- Headers:
  - `Content-Type`: `video/mp4` / `video/mp2t` / `application/vnd.apple.mpegurl`
  - `Content-Length`: 文件大小(字节)
  - `ETag`: `"{MD5值}"`
  - `content-md5`: MD5哈希值
  - `Last-Modified`: RFC1123格式时间戳
  - `Accept-Ranges`: `bytes` (支持范围请求)
  - `Cache-Control`: `public,max-age=31536000`
  - `Vary`: `Origin`
  - CORS相关头:
    - `Access-Control-Allow-Origin: *`
    - `Access-Control-Allow-Methods: GET, OPTIONS`
    - `Access-Control-Allow-Headers: Range, Origin, Content-Type`
    - `Access-Control-Expose-Headers: Content-Range, Last-Modified, Etag, content-md5, Content-Length, Accept-Ranges, Vary`
- Body: 视频文件二进制流

**错误响应:**

| 状态码 | 响应体示例 | 说明 |
|--------|------------|------|
| 401 | `{"code":401,"msg":"IMEI validation not pass"}` | IMEI鉴权失败 |
| 400 | `{"code":1,"msg":"文件不存在"}` | 文件获取失败 |
| 500 | `{"code":-1,"msg":"IMessage:服务内部错误"}` | 服务内部错误 |

#### 4.2.2 文件下载接口

**端点:** `GET /download/*`

**功能说明:** 提供文件下载功能，模拟MUEN媒体服务器的下载接口。

**请求参数:**

| 参数名 | 位置 | 类型 | 必填 | 描述 |
|--------|------|------|------|------|
| `:splat` | 路径参数 | string | 是 | 下载文件路径 |

**处理流程:**
1. 解析路径参数获取文件路径
2. 调用VideoService.Download()读取本地文件
3. 设置文件下载响应头
4. 将文件内容写入响应流

**成功响应:**
- HTTP状态码: `200 OK`
- Headers:
  - `Content-Disposition`: `attachment; filename={filename}`
  - `Content-Length`: 文件大小
- Body: 文件二进制流
- JSON响应(下载完成返回): 
```json
{
  "code": 200,
  "msg": "download success",
  "data": true
}
```

**错误响应:**
- HTTP状态码: `400`
- 响应体: `{"code":1,"msg":"错误描述"}`

#### 4.2.3 健康检查接口

**端点:** `GET /test`

**功能说明:** 用于服务健康检查和连接性测试。

**请求参数:** 无

**成功响应:**
- HTTP状态码: `200 OK`
- 响应体:
```json
{
  "code": 0,
  "msg": "test success",
  "data": true
}
```

### 4.3 API版本和兼容性

**版本策略:** 
- 当前无版本前缀，保持现有路径不变
- 未来扩展可在路径前添加`/api/v{version}/`

**向后兼容保证:**
- 所有现有API路径、方法、参数名称、参数类型保持不变
- 响应结构(BaseResponse、DataResponse)字段名和JSON标签保持不变
- 响应码定义(retcode常量)保持不变

---

## 5. 数据模型设计

### 5.1 核心数据模型（必须保持字段一致）

#### 5.1.1 基础响应模型
```go
// 基础响应结构
type BaseResponse struct {
    Code    int    `json:"code"`    // 响应状态码: 0=成功, 401=鉴权失败, -1=内部错误
    Message string `json:"msg"`     // 响应消息描述
}
```

#### 5.1.2 数据响应模型
```go
// 带数据的响应结构
type DataResponse struct {
    BaseResponse
    Data interface{} `json:"data"`  // 实际数据，可以是任意类型
}
```

#### 5.1.3 文件信息模型
```go
// 文件元数据模型
type FileInfo struct {
    Name             string              // 文件名
    Path             string              // 文件相对路径
    Size             string              // 文件大小(字符串格式)
    ModifiedTime     time.Time           // 最后修改时间
    Hash             string              // 文件内容的MD5哈希值(十六进制字符串)
    HasCached        bool                // 是否已缓存到本地
    ExtraWriteTarget io.WriteSeeker      // 额外写入目标(可选，用于同时写入多个地方)
    Finalizer        func()              // 资源清理回调函数
}

// 添加最终清理器方法(保持一致性)
func (f FileInfo) AddFinalizer(newFinalizer func()) {
    oldFinalizer := f.Finalizer
    f.Finalizer = func() {
        if oldFinalizer != nil {
            oldFinalizer()
        }
        newFinalizer()
    }
}
```

#### 5.1.4 告警相关模型
```go
// 告警事件模型
type AlarmEvent struct {
    AlarmID      string                   // 告警ID
    EventMessage string                   // 事件详细信息
    Type         base.GenerateOrClearType // 事件类型: 生成告警/清除告警
}

// 告警参数信息
type AlarmParamInfo struct {
    ParamName  string `json:"paramName"`  // 参数名
    ParamValue string `json:"paramValue"` // 参数值
}

// 告警信息结构
type AlarmInfo struct {
    Location   string `json:"location,omitempty"`   // 位置信息
    AppendInfo string `json:"appendInfo,omitempty"` // 附加信息
    AlarmId    string `json:"alarmId,omitempty"`    // 告警ID
}

// 告警响应结构
type AlarmResponse struct {
    Retdesc string      `json:"retdesc,omitempty"` // 返回描述
    Data    []AlarmInfo `json:"data,omitempty"`     // 告警数据列表
    RetCode string      `json:"retcode,omitempty"`  // 返回码
}
```

### 5.2 内部数据模型

#### 5.2.1 配置中心配置项
```go
type ConfigValue struct {
    Value string // 配置值
    // 可扩展字段: 版本、时间戳等
}
```

#### 5.2.2 任务执行状态
```go
type TaskResult struct {
    TaskName    string    // 任务名称
    ExecutedAt  time.Time // 执行时间
    Status      string    // 状态: success/failed
    Error       error     // 错误信息
    Description string    // 执行描述
}
```

### 5.3 数据流转模型

#### 视频请求处理流程数据流:
```
客户端请求 → Controller → AuthService验证 → VideoService处理 → Storage检查
                                                              ↓
                    MUEN远程获取 ← Remote ← 缓存未命中 ─┴ 缓存命中
                                                              ↓
                                                     FileInfo元数据生成
                                                              ↓
                                                    HTTP响应头设置+文件流
```

---

## 6. 配置依赖设计

### 6.1 配置层次结构

```
┌─────────────────────────────────────┐
│        环境变量层                     │  APPID, PODNAME, MUEN_MEDIA_URL_PREFIX等
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│        配置中心层                     │  CSE动态配置: moon::mediaEndpoint等
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│        本地配置层                     │  conf.Config结构体
│   - 编译时常量                       │
│   - 默认值配置                        │
│   - 文件配置(如果有)                  │
└─────────────────────────────────────┘
```

### 6.2 配置项详细说明

#### 6.2.1 基础服务配置

| 配置项 | 类型 | 默认值 | 配置方式 | 说明 |
|--------|------|--------|----------|------|
| `MediaCache` | string | `/opt/mtuser/mcs/video` | 代码/环境变量 | 本地缓存存储路径 |
| `HTTPTimeout` | int | `600` | 代码 | HTTP请求超时时间(秒) |
| `httpport` | int | `9996` | beego.AppConfig | 内部HTTP服务端口 |
| `httpsport` | int | `9997` | beego.AppConfig | 内部HTTPS服务端口 |
| `moon::httpport` | int | `9990` | 配置中心 | 外部HTTP服务端口 |
| `moon::httpsport` | int | `9991` | 配置中心 | 外部HTTPS服务端口 |

#### 6.2.2 日志配置

| 配置项 | 结构 | 默认值 | 说明 |
|--------|------|--------|------|
| `Logger.LogFile` | string | `/opt/mtuser/mcs/log/log1` | 日志文件路径 |
| `Logger.LogLevel` | string | `INFO` | 日志级别: INFO/WARN/DEBUG |

#### 6.2.3 数据老化配置

| 配置项 | 类型 | 默认值 | 环境变量 | 说明 |
|--------|------|--------|----------|------|
| `ScanningTaskPeriod` | string(小时) | `1` | `SCANNING_TASK_PERIOD` | 缓存扫描周期 |
| `ClearingTaskPeriod` | string(小时) | `24` | `CLEARING_TASK_PERIOD` | 文件清理周期 |
| `ClearingTaskThreshold` | string(MB) | `491520`(480GB) | `CLEARING_TASK_THRESHOLD` | 触发清理的磁盘阈值 |
| `FileAccessInactiveThreshold` | string(天) | `10` | `FILE_ACCESS_INACTIVE_THRESHOLD` | 文件未活跃阈值 |
| `DeleteInactiveFileTimeout` | string(秒) | `60` | `DELETE_INACTIVE_FILE_TIMEOUT` | 删除文件超时 |
| `CacheAvailable` | bool | `false` | 运行时设置 | 缓存可用标志 |

#### 6.2.4 远程服务配置

| 配置项 | 类型 | 配置方式 | 说明 |
|--------|------|----------|------|
| `MUEN_MEDIA_URL_PREFIX` | string | 环境变量 | MUEN媒体服务地址前缀 |
| `moon::mediaEndpoint` | string | 配置中心 | MUEN HTTP媒体端点 |
| `moon::httpsMediaEndpoint` | string | 配置中心 | MUEN HTTPS媒体端点 |
| `moon::enableHttps` | string | 配置中心 | 是否启用HTTPS (`true`/`false`) |

#### 6.2.5 CSE配置

| 配置项 | 类型 | 环境变量 | 说明 |
|--------|------|----------|------|
| `APPID` | string | `APPID` | 应用ID |
| `PODNAME` | string | `PODNAME` | Pod名称 |
| `NAMESPACE` | string | `NAMESPACE` | 命名空间 |
| `NODENAME` | string | `NODENAME` | 节点名称 |
| `SERVICENAME` | string | `SERVICENAME` | 服务名称 |
| `ENABLE_HTTP` | string | `ENABLE_HTTP` | 是否启用HTTP服务 (`true`/`false`) |

#### 6.2.6 网络接口配置

| 网络标识符 | 网络接口名称 | 用途 |
|-----------|--------------|------|
| `FABRIC_ETH` | `bond-base` | 内部网络接口(检索IP用) |
| `SC_TRUNK_ETH` | `bond-external` | 外部网络接口(检索IP用) |

### 6.3 配置加载顺序

```
1. 初始化默认值 (config.go init函数)
      ↓
2. 环境变量覆盖 (getEnv函数)
      ↓
3. 命令行参数解析 (flagutil.Parse)
      ↓
4. 配置中心动态获取 (运行时)
      ↓
5. 应用启动时读取实例属性
```

### 6.4 配置依赖关系

```
MediaCache (缓存路径)
    ↓
影响: storage.local_storage.go 中的文件读写操作

CacheAvailable (缓存可用状态)
    ↓
影响: VideoService.GetVideo() 中的缓存逻辑

DataAging 配置
    ↓
影响: tasks/scanning_task.go 和 tasks/clearing_task.go

GSF/CSE配置
    ↓
影响: 服务注册、发现、配置中心获取

网络接口配置
    ↓
影响: https.GetLocalIP() 获取服务监听地址
```

### 6.5 热更新配置支持

当前系统支持以下配置的热更新:
- `moon::mediaEndpoint` (通过配置中心)
- `moon::httpsMediaEndpoint` (通过配置中心)
- `moon::enableHttps` (通过配置中心)

其他配置项需要重启服务生效，未来可扩展支持更多动态配置项。

---

## 7. 运行时依赖设计

### 7.1 外部服务依赖

| 服务名称 | 类型 | 依赖方式 | 可用性处理 | 重试策略 |
|---------|------|----------|-------------|----------|
| MUEN Media Server | HTTP/HTTPS | 配置中心/环境变量 | 告警+降级 | 无(直接失败) |
| GIDS Authentication | HTTP | 通过CSE服务发现 | 鉴权失败返回401 | 无(直接失败) |
| CSE Service Registry | gRPC | GSF框架集成 | 服务注册失败重试 | 初始化时重试360次 |
| CSE Config Center | gRPC | 框架集成 | 使用默认值 | 框架处理 |
| Alarm SDK (FM) | HTTP | CSP SDK集成 | 告警失败重试+超时 | 3次重试，等待10秒 |

### 7.2 依赖可用性策略

#### 7.2.1 服务启动依赖链
```
1. GSF框架初始化 (最多重试360次，每次等待5秒)
   ↓ (成功)
2. NTP服务初始化
   ↓
3. Runlog和ModuleKeeper初始化
   ↓
4. 定时任务初始化
   ↓
5. 远程客户端初始化 (MuenClient)
   ↓
6. 启动HTTP/HTTPS服务器
   ↓
7. 清除历史告警
   ↓ (完成) 服务就绪
```

#### 7.2.2 运行时依赖处理策略

**MUEN服务不可用:**
- 触发告警 (AlarmId300020)
- 返回错误响应给客户端
- 告警抑制: 10分钟内重复告警不发送

**GIDS服务不可用:**
- 鉴权接口返回错误
- 客户端收到401或500错误
- 不触发告警

**CSE注册失败:**
- 启动阶段: 重试360次后退出
- 运行时: 通过goroutine定期获取实例信息

**告警服务不可用:**
- 告警静默失败
- 记录错误日志
- 不影响主业务流程

### 7.3 依赖健康检查

| 依赖项 | 检查方式 | 检查频率 | 故障检测时间 |
|--------|----------|----------|--------------|
| GSF框架 | 框架内部健康检查 | 连续启动时 | 30分钟逐次检查 |
| 远程服务 | HTTP响应状态 | 每次请求 | 实时 |
| CSE实例 | 定期查询 | 每10秒 | 两次查询周期 |
| 本地磁盘 | 扫描任务 | 每小时 | 1小时 |

---

## 8.设计 安全

### 8.1 认证授权机制

#### IMEI设备认证
```
客户端请求
    ↓
VideoController.GetVideo()
    ↓
AuthService.ValidateIMEI(imei, checkType)
    ↓
remote.PostValidateIMEI() → GIDS服务
    ↓
返回验证结果 → 允许/拒绝访问
```

#### 鉴权状态码映射
| GIDS返回状态码 | 本地处理 | 最终返回客户端 |
|---------------|----------|----------------|
| 200 | 鉴权通过 | 正常执行业务逻辑 |
| 401 | 鉴权失败 | 返回401响应，IMEI无效 |
| 500 | 服务错误 | 返回500响应，服务内部错误 |
| 其他 | 未知状态 | 返回错误，状态码异常 |

### 8.2 网络安全

#### HTTPS支持
- 外部HTTPS服务运行在端口9991
- 证书通过`cert`包动态加载
- 支持证书热更新订阅

#### 网络隔离
- 内部服务: bind到 `FABRIC_ETH` (`bond-base`) 接口，端口9996/9997
- 外部服务: bind到 `SC_TRUNK_ETH` (`bond-external`) 接口，端口9990/9991

### 8.3 访问控制

#### 过载防护 (Greatwall)
- 全局过滤器 `OverLoadFilter`
- 基于 `APIService` 维度的访问限流
- 超限返回`429 Too Many Requests`
- 建议 `Retry-After: 3` 秒

#### 文件路径安全
- 使用 `filepath.Clean()` 清理路径
- 防止路径穿越攻击
- 文件扩展名白名单: `.ts`, `.m3u8`, `.mp4`

### 8.4 数据安全

#### 响应头CORS配置
```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: false
Access-Control-Allow-Methods: GET, OPTIONS
Access-Control-Allow-Headers: Range, Origin, Content-Type
Access-Control-Expose-Headers: Content-Range, Last-Modified, Etag, content-md5, Content-Length, Accept-Ranges, Vary
```

---

## 9. 监控与告警设计

### 9.1 日志记录策略

| 模块 | 级别 | 记录内容 | 示例 |
|------|------|----------|------|
| VideoController | INFO | 请求处理时间 | `file[xxx] use time: 123ms` |
| VideoController | ERROR | 文件获取失败 | `get video failed, err: xxx` |
| AuthService | INFO | 鉴权开始/完成 | `Start to validate IMEI`, `IMEI validation completed, result: true` |
| AuthService | ERROR | 鉴权失败 | `IMEI validation failed: xxx` |
| Tasks | INFO | 任务执行 | `success scanning dir size, current cache available stat is: xxx` |
| AlarmService | INFO | 告警发送/清除 | `Alarm xxx send successfully` |

### 9.2 告警设计

#### 告警类型
| 告警ID | 触发条件 | 清除条件 | 抑制时间 |
|--------|----------|----------|----------|
| 300020 | 从MUEN获取视频失败 | 成功获取到视频 | 10分钟 |

#### 告警机制
- 异步告警事件通道 (`alarmEventChanel`), 容量999
- 10分钟重复告警抑制
- 服务启动时清除本节点历史告警
- 与FM服务集成获取/清除活动告警

### 9.3 性能指标

| 指标 | 测量方式 | 理想范围 |
|------|----------|----------|
| 视频请求响应时间 | 每次请求记录 | < 1s (缓存命中) |
| 鉴权响应时间 | 鉴权服务调用 | < 500ms |
| 缓存命中率 | 服务运行统计 | > 95% |
| 磁盘使用率 | 定时扫描任务 | < 85% |

---

## 10. 扩展性设计建议

### 10.1 存储层扩展

当前仅支持本地文件存储，可扩展为:
- S3兼容对象存储 (MinIO, AWS S3)
- Redis缓存层 (热点文件)
- 多级缓存 (内存 -> 本地磁盘 -> 对象存储)

### 10.2 服务发现扩展

当前仅支持GIDS通过CSE发现，可扩展:
- 支持多认证服务提供方
- 增加本地缓存认证结果
- 提供备用认证服务

### 10.3 配置扩展

建议增加:
- 配置文件支持 (YAML/TOML)
- 配置验证和热更新
- 配置版本管理和回滚

### 10.4 监控扩展

建议增加:
- Prometheus metrics暴露
- 链路追踪集成 (Jaeger/Zipkin)
- 实时监控面板 (Grafana)

---

## 11. 重构建议

### 11.1 保持兼容性

重构时必须保持:
✅ 所有API路径签名不变
✅ 请求/响应结构字段名和JSON标签不变
✅ 错误码和错误消息格式不变
✅ 配置项名称和语义不变
✅ 环境变量名称不变

### 11.2 优先重构领域

| 优先级 | 模块 | 原因 | 建议 |
|--------|------|------|------|
| 高 | storage层 | 接口设计好但实现单一 | 抽象存储策略，支持多种后端 |
| 高 | service层 | 业务逻辑和外部耦合 | 进一步解耦，增加单元测试覆盖 |
| 中 | remote层 | 错误处理不够完善 | 增加重试、熔断、降级机制 |
| 中 | config层 | 配置加载分散 | 统一配置管理，增加验证 |
| 低 | common层 | 工具函数较简单 | 按职责进一步分类 |

### 11.3 测试策略

重构过程中建议:
- 为每个模块编写单元测试
- 集成测试覆盖关键API路径
- 性能测试确保不退化
- 兼容性测试保证现有客户端正常工作

---

## 12. 总结

MediaCacheService是一个设计良好的媒体缓存微服务，具备以下特点:

**优势:**
- 分层清晰，职责明确
- 接口抽象良好，便于扩展
- 集成华为CSP微服务生态
- 支持多端口、多网卡部署
- 具备基础监控和告警能力

**改进空间:**
- 存储层可支持多种后端
- 远程服务调用可增加可靠性机制
- 配置管理可更加灵活
- 测试覆盖度有待提升

**重构核心原则:**
- 保持所有现有API接口签名不变
- 保持所有数据结构和字段定义不变
- 在保证兼容性的前提下，优化内部实现
- 增加测试覆盖，确保重构安全

本文档可作为重构的指导参考，确保重构过程中不破坏现有功能和接口兼容性。