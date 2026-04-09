# MediaCache Service 模块分析文档

## 目录概述

本文档详细分析了 `D:\CloudCellular\MediaCacheService\src\service` 目录下的代码结构，包含三个主要的Go文件：

- `AlarmService.go` - 告警服务模块
- `VideoService.go` - 视频服务模块  
- `auth_service.go` - 鉴权服务模块

## 目录层级关系

```
service/
├── AlarmService.go     (11,656 bytes)
├── VideoService.go     (2,721 bytes)
└── auth_service.go     (1,117 bytes)
```

## 接口、结构体、函数分析

### 1. AlarmService.go

#### 接口

**AlarmService** - 告警服务接口
```go
type AlarmService interface {
    SendAlarm(alarmID, EventMessage string)
    ClearAlarm(alarmID, EventMessage string)
}
```

#### 结构体

**AlarmEvent** - 告警事件结构体
```go
type AlarmEvent struct {
    AlarmID      string                   // 告警ID
    EventMessage string                   // 事件信息
    Type         base.GenerateOrClearType // 事件类型
}
```

**AlarmServiceImpl** - 告警服务实现结构体
```go
type alarmServiceImpl struct {
    mu           sync.Mutex               // 互斥锁，线程安全
    alarms       map[string]int64         // 告警记录表: 告警ID -> 上报时间戳
    alarmManager base.CSPAlarmManager     // 告警管理器
}
```

**AlarmParamInfo** - 告警参数信息
```go
type AlarmParamInfo struct {
    ParamName  string `json:"paramName"`  // 告警参数名
    ParamValue string `json:"paramValue"` // 告警参数值
}
```

**AlarmInfo** - 告警信息
```go
type AlarmInfo struct {
    Location   string `json:"location,omitempty"`
    AppendInfo string `json:"appendInfo,omitempty"`
    AlarmId    string `json:"alarmId,omitempty"`
}
```

**AlarmResponse** - 告警响应
```go
type AlarmResponse struct {
    Retdesc string      `json:"retdesc,omitempty"`
    Data    []AlarmInfo `json:"data,omitempty"`
    RetCode string      `json:"retcode,omitempty"`
}
```

#### 函数实现详解

**初始化相关函数:**

- `NewAlarmService() AlarmService`
  - 功能：创建告警服务实例
  - 实现：返回全局的 alarmService 实例
  - 返回类型：接口 AlarmService

- `init()`
  - 功能：包初始化函数，在main函数执行前调用
  - 实现：
    - 初始化 alarmService 结构体，设置告警管理器
    - 创建告警事件通道 (alarmEventChanel，容量为999)
    - 注册告警管理器
    - 启动告警事件处理协程

**核心业务函数:**

- `SendAlarm(alarmID, EventMessage string)`
  - 功能：发送告警
  - 实现：
    - 创建告警事件对象，类型为 GenerateAlarm
    - 通过通道发送告警事件供协程处理
  - 调用链：调用者 → sendAlarmEvent() → 通道 → handleEvent() → sendAlarm()

- `ClearAlarm(alarmID, eventMessage string)`
  - 功能：清除告警
  - 实现：
    - 创建告警事件对象，类型为 ClearAlarm
    - 通过通道发送告警事件供协程处理
  - 调用链：调用者 → sendAlarmEvent() → 通道 → handleEvent() → clearAlarm()

**内部实现函数:**

- `handleEvent()` (go routine)
  - 功能：告警事件处理协程，持续运行
  - 实现：无限循环监听通道，根据事件类型调用相应的处理函数
  - 模式：生产者-消费者模式，异步处理告警

- `sendAlarm(event AlarmEvent) bool`
  - 功能：具体发送告警逻辑
  - 实现：
    - 检查同一告警10分钟内的重复上报限制
    - 调用 reportAlarm() 发送告警
    - 更新告警记录表中的上报时间戳
  - 返回：发送成功返回true，失败返回false

- `clearAlarm(alarmEvent AlarmEvent)`
  - 功能：具体清除告警逻辑
  - 实现：
    - 检查告警是否存在，不存在则跳过
    - 调用 reportAlarm() 发送清除告警
    - 从记录表中删除该告警

- `reportAlarm(alarmEvent AlarmEvent) bool`
  - 功能：向告警系统上报告警
  - 实现：
    - 创建告警对象
    - 添加告警参数（kind、namespace、sourceip等）
    - 最多重试2次，每次间隔10秒
  - 返回：上报成功返回true，失败返回false

**公共函数:**

- `CleanAllActiveAlarm() bool`
  - 功能：清除所有活动告警
  - 实现：
    - 从 FMService 获取告警信息（最多重试360次）
    - 根据源IP过滤本节点的告警
    - 调用 clearHistoryAlarm() 清除每个告警
  - 返回：清除成功返回true，否则返回false

**工具函数:**

- `GetAllActiveAlarmFromFMService(alarmIds string) (map[string][]AlarmParamInfo, error)`
- `handlerActivityAlarmData(micServiceName string, requestParams []byte) (map[string][]AlarmParamInfo, error)`
- `OSHttpsGetRequestByCSE(url string, microServiceName string, method string, body []byte) (string, error)`

#### 常量配置

```go
const (
    ValuesLen       = 2                      // 位置信息数组长度
    TimePeriodInit  = 3                      // 初始化超时时间（秒）
    TimePeriodClean = 5                      // 清理超时时间（秒）
    RetryTimes      = 360                    // 重试次数
    AlarmId300020   = "300020"              // 告警ID常量
    maxAlarmListLen = 999                    // 告警列表最大长度
    RespOK          = 200                    // HTTP响应码正常
    POST            = "POST"                 // HTTP POST方法
    GetActiveAlarms = "GET_ACTIVE_ALARMS"    // 获取活动告警命令
    ResultCode      = "0"                    // 结果码（0表示正常）
)
```

### 2. VideoService.go

#### 接口

**VideoService** - 视频服务接口
```go
type VideoService interface {
    GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
    Download(path string) (io.ReadCloser, int64, error)
}
```

#### 结构体

**VideoServiceImpl** - 视频服务实现结构体
```go
type VideoServiceImpl struct {
    remote  remote.Remote    // 远程服务接口
    storage storage.Storage  // 存储服务接口
    alarm   AlarmService     // 告警服务接口
}
```

#### 函数实现详解

**初始化函数:**

- `NewVideoService() *VideoServiceImpl`
  - 功能：创建视频服务实例
  - 实现：依赖注入，设置remote、storage、alarm三个依赖项
  - 返回：VideoServiceImpl 指针

**核心业务函数:**

- `GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)`
  - 功能：获取视频文件，支持本地缓存和远程下载
  - 实现逻辑：
    1. 清理视频路径
    2. 检查本地缓存是否存在
    3. 如果缓存存在，直接从本地读取
    4. 如果缓存不存在，从远程MUEN服务下载
    5. 如果下载成功，将文件保存到本地缓存
    6. 对于沐恩获取失败的情况，发送告警
    7. 如果缓存告警存在，则清除告警
  - 返回：读取流、文件信息、错误信息

- `Download(videoPath string) (io.ReadCloser, int64, error)`
  - 功能：从文件系统直接下载视频
  - 实现逻辑：
    1. 检查文件是否存在
    2. 打开文件并返回文件流和文件大小
  - 注意：此函数直接操作文件系统，不经过缓存机制
  - 返回：文件读取流、文件大小、错误信息

#### 函数调用关系

```
NewVideoService()
└── 创建 VideoServiceImpl 实例
    ├── remote.NewRemoteImpl()    <!-- 创建远程服务实例 -->
    ├── storage.NewLocalStorage() <!-- 创建存储服务实例 -->
    └── NewAlarmService()        <!-- 创建告警服务实例 -->

GetVideo()
└── 本地缓存检查
    ├── storage.Exist(videoPath)      <!-- 检查文件是否存在 -->
    ├── storage.Get(videoPath)        <!-- 从缓存读取文件 -->
└── 远程下载处理
    ├── remote.GetVideo(videoPath)    <!-- 从MUEN远程获取文件 -->
    ├── alarm.SendAlarm()             <!-- 发送告警 -->
    ├── alarm.ClearAlarm()            <!-- 清除告警 -->
    └── storage.Cache(videoPath)      <!-- 缓存文件到本地 -->
```

### 3. auth_service.go

#### 接口

**AuthService** - 鉴权服务接口
```go
type AuthService interface {
    ValidateIMEI(imei string, checkType string) (bool, error)
}
```

#### 结构体

**AuthServiceImpl** - 鉴权服务实现结构体
```go
type AuthServiceImpl struct {
    remote remote.Remote // 远程服务接口
}
```

#### 函数实现详解

**初始化函数:**

- `NewAuthService() *AuthServiceImpl`
  - 功能：创建鉴权服务实例
  - 实现：依赖注入，设置remote依赖项
  - 返回：AuthServiceImpl 指针

**核心业务函数:**

- `ValidateIMEI(imei string, checkType string) (bool, error)`
  - 功能：验证IMEI是否有效
  - 实现逻辑：
    1. 记录开始时间和操作日志
    2. 调用远程服务的PostValidateIMEI方法进行验证
    3. 处理异常情况，记录错误日志
    4. 记录完成时间和验证结果
    5. 返回验证结果和可能的错误
  - 返回：验证结果（有效/无效）、错误信息

## 模块间依赖关系

### 1. 依赖关系图

```
service/
├── AlarmService.go      (独立模块，无内部依赖)
├── VideoService.go        
│   └── 依赖: remote, storage, AlarmService
└── auth_service.go      
    └── 依赖: remote
```

### 2. 外部依赖分析

**AlarmService.go** 外部依赖：
- ` AlarmSDK_GO/api/alarmapi` - 告警SDK API
- ` AlarmSDK_GO/api/base` - 告警基础类型
- ` Go-chassis-extend/api/ServiceComb/go-chassis/...` - 微服务框架依赖
- ` MediaCacheService/common/constants` - 常量定义
- ` MediaCacheService/common/logger` - 日志服务
- ` context`, `encoding/json`, `errors`, `os`, `sync`, `time` - Go标准库

**VideoService.go** 外部依赖：
- ` MediaCacheService/common/logger` - 日志服务
- ` MediaCacheService/remote` - 远程服务
- ` MediaCacheService/storage` - 存储服务
- ` fmt`, `io`, `os`, `filepath`, `time` - Go标准库

**auth_service.go** 外部依赖：
- ` MediaCacheService/common/logger` - 日志服务
- ` MediaCacheService/remote` - 远程服务
- ` fmt`, `time` - Go标准库

## 服务协调关系

### 1. 服务间调用关系

```
VideoService.GetVideo()
├── 检查本地缓存
└── 缓存不存在时
    ├── remote.GetVideo()     <!-- 调用远程服务 -->
    ├── alarm.SendAlarm()     <!-- 发送失败告警 -->
    └── alarm.ClearAlarm()    <!-- 发送成功清除告警 -->

AuthServiceImpl.ValidateIMEI()
└── remote.PostValidateIMEI() <!-- 调用远程鉴权服务 -->
```

### 2. 告警系统集成

视频服务与告警系统紧密集成：
- MUEN服务获取失败时触发告警 (AlarmId300020)
- 成功获取后自动清除对应的告警
- 告警系统支持重复上报限制（10分钟内不重复）

## 设计模式

### 1. 接口分离原则
每个服务都明确定义了接口，实现了接口与实现的分离。

### 2. 依赖注入模式
VideoServiceImpl 和 AuthServiceImpl 通过构造函数注入依赖，便于测试和解耦。

### 3. 生产者-消费者模式
AlarmService采用异步处理，通过channel实现告警事件的异步处理，避免阻塞业务流程。

### 4. 委托模式
VideoService委托 remote.Remote 和 storage.Storage 进行实际的文件获取和存储操作。

### 5. 策略模式
通过接口定义，可以灵活替换不同的实现策略（如不同的存储后端、远程服务等）。

## 异常处理

### 1. 错误重试机制
- 告警服务支持重试机制（最多2次，每次间隔10秒）
- 清除活动告警支持最大重试360次

### 2. 异常阻塞避免
- 告警通道设置超时机制（5秒），避免阻塞主流程
- 远程调用设置超时控制

### 3. 防重复告警
- 告警系统记录告警ID和上报时间
- 10分钟内的重复同类型告警会被过滤掉

## 线程安全

### 1. 互斥锁保护
- AlarmServiceImpl 使用 sync.Mutex 保证告警记录表的线程安全
- GetActiveAlarmFromFMService 使用 RWMutex 保证并发安全

### 2. Channel通信
- 告警事件通过Channel传递，避免共享内存竞争
- 支持异步处理，提高系统吞吐量

## 性能优化

### 1. 缓存机制
- VideoService采用本地缓存，减少远程调用
- 支持文件信息缓存，避免重复读取

### 2. 异步处理
- 告警采用异步处理，避免影响主业务流程
- 支持批量处理和能力限制

### 3. 连接池
- HTTP客户端使用连接池配置
- 设置最大空闲连接数和超时时间

## 扩展性设计

### 1. 接口标准化
所有服务都定义了清晰的接口，便于扩展和替换实现。

### 2. 配置外部化
服务行为通过配置文件和外部变量控制，便于动态调整。

### 3. 模块化设计
各服务模块职责单一，易于维护和扩展。

## 总结

service模块是MediaCacheService业务逻辑的核心，包含三个主要服务：

1. **AlarmService** - 负责系统告警管理，采用异步处理模式，支持告警的发送、清除和历史管理
2. **VideoService** - 负责视频文件的获取和缓存，结合本地缓存和远程下载，集成告警系统
3. **AuthService** - 负责设备鉴权，通过远程服务验证IMEI有效性

三个服务通过接口定义实现松耦合，支持灵活的依赖注入和替换。系统采用了多种设计模式和优化策略，确保了高性能、高可用和良好的扩展性。