# MediaCacheService 路由系统分析文档

## 目录
1. [概述](#概述)
2. [接口分析](#接口分析)
3. [结构体分析](#结构体分析)
4. [函数实现详解](#函数实现详解)
5. [调用关系图](#调用关系图)
6. [HTTP API接口](#http-api接口)
7. [路由注册机制](#路由注册机制)
8. [中间件机制](#中间件机制)

## 概述

MediaCacheService 是一个基于 Go 语言实现的高性能媒体缓存服务，使用 Beego v2 作为 Web 框架，同时支持 Go-Chassis 微服务框架。该服务提供视频文件的缓存、流媒体播放、文件下载等功能，并包含完整的认证、监控和管理功能。

### 技术栈
- **Go 版本**: 1.20
- **Web 框架**: Beego v2.1.0
- **微服务框架**: Go-Chassis-extend
- **华为企业SDK**: CSPGSOMF、AlarmSDK、GreatWall SDK
- **网络库**: Gorilla WebSocket、Redigo (Redis)

## 接口分析

### 1. 路由相关的核心接口

#### 1.1 IController 接口

**文件位置**: `src/controllers/controller.go`

```go
type IController interface {
    beego.ControllerInterface
    RouteInfo() RouteInfo
}
```

**接口说明**:
- 继承自 Beego 的 `ControllerInterface`
- 要求所有控制器实现 `RouteInfo()` 方法用于路由信息注册
- 定义了控制器的基本契约

#### 1.2 RouteInfo 结构体

**文件位置**: `src/controllers/controller.go`

```go
type RouteInfo struct {
    RouteMapping map[string]string      // 路径到方法的映射
    Filters      map[FilterAction]beego.FilterFunc  // 中间件过滤器
}
```

**字段说明**:
- `RouteMapping`: HTTP 路径到控制器方法的映射，格式为 `"路径:HTTP方法:方法名"`
- `Filters`: 过滤器映射，支持 `Before` 和 `After` 两种类型的过滤器

#### 1.3 AuthService 接口

**文件位置**: `src/service/auth_service.go`

```go
type AuthService interface {
    ValidateIMEI(imei string, checkType string) (bool, error)
}
```

**接口说明**:
- 负责验证设备 IMEI 的合法性
- 返回验证结果和可能的错误

#### 1.4 VideoService 接口

**文件位置**: `src/service/VideoService.go`

```go
type VideoService interface {
    GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
    Download(path string) (io.ReadCloser, int64, error)
}
```

**接口说明**:
- `GetVideo`: 获取视频流和文件信息，支持缓存读取
- `Download`: 文件下载功能，返回文件读取器和文件大小

#### 1.5 Storage 接口

**文件位置**: `src/storage/storage.go`

```go
type Storage interface {
    Cache(filePath string) (*FileInfo, error)
    Get(videoPath string) (io.ReadCloser, *FileInfo, error)
    Exist(filePath string) bool
}
```

**接口说明**:
- `Cache`: 将文件缓存到存储中
- `Get`: 从存储中获取文件
- `Exist`: 检查文件是否存在于存储中

### 1.6 Go-Chassis 相关接口

#### API 结构体

**文件位置**: `src/routers/router.go`

```go
type API struct {
    // 实现 Go-Chassis RESTful API 接口
}
```

**接口实现**:
- `URLPatterns()` 方法返回路由定义
- `GetVideo(c rf.Context)`: 处理视频请求
- `Download(c rf.Context)`: 处理下载请求
- `Test(c rf.Context)`: 测试接口

## 结构体分析

### 1. BaseController 结构体

**文件位置**: `src/controllers/controller.go`

```go
type BaseController struct {
    beego.Controller
}
```

**功能说明**:
- 提供所有控制器的基础功能
- 封装了通用的 HTTP 请求处理
- 提供响应方法和工具函数

**方法解析**:

#### 1.1 基础方法

```go
func (c *BaseController) QueryParameter(name string) string
```
- **功能**: 获取 HTTP 请求查询参数
- **参数**: `name` - 参数名称
- **返回**: 参数值

```go
func (c *BaseController) PathParameter(name string) string
```
- **功能**: 获取 URL 路径参数
- **参数**: `name` - 参数名称
- **返回**: 路径参数值

#### 1.2 响应方法

```go
func (c *BaseController) OK(data interface{})
```
- **功能**: 返回成功的 JSON 响应
- **参数**: `data` - 响应数据
- **状态码**: 200

```go
func (c *BaseController) Failed(data resp.BaseResponse)
```
- **功能**: 返回失败的 JSON 响应
- **参数**: `data` - 响应数据
- **状态码**: 400

```go
func (c *BaseController) InternalServiceError()
```
- **功能**: 返回内部服务错误响应
- **状态码**: 500

#### 1.3 辅助方法

```go
func (c *BaseController) WriteHeaderAndJSON(status int, v interface{}, contentType string) error
```
- **功能**: 设置 HTTP 响应头并返回 JSON 数据
- **参数**: 
  - `status`: HTTP 状态码
  - `v`: 要序列化的数据
  - `contentType`: Content-Type 头值

```go
func (c *BaseController) DownloadFile(file string)
```
- **功能**: 下载文件
- **参数**: `file` - 文件路径
- **功能**: 设置适当的 HTTP 头后发送文件

### 2. VideoController 结构体

**文件位置**: `src/controllers/video_controller.go`

```go
type VideoController struct {
    BaseController
    videoService service.VideoService
    authService  service.AuthService
}
```

**功能说明**:
- 处理视频相关的 HTTP 请求
- 集成了认证和视频服务
- 支持视频流和文件下载

### 3. AuthServiceImpl 结构体

**文件位置**: `src/service/auth_service.go`

```go
type AuthServiceImpl struct {
    remote remote.Remote
}
```

**功能说明**:
- 实现 IMEI 认证服务的具体实现
- 通过远程服务验证设备合法性
- 记录认证状态和耗时

### 4. VideoServiceImpl 结构体

**文件位置**: `src/service/VideoService.go`

```go
type VideoServiceImpl struct {
    remote  remote.Remote
    storage storage.Storage
    alarm   AlarmService
}
```

**功能说明**:
- 视频服务的具体实现
- 管理视频缓存逻辑
- 处理远程服务获取和本地缓存
- 集成告警功能

### 5. FileInfo 结构体

**文件位置**: `src/storage/storage.go`

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

**功能说明**:
- 文件信息的封装
- 支持文件元数据和缓存状态
- 提供文件清理函数机制

## 函数实现详解

### 1. 路由相关函数

#### 1.1 URLPatterns() 函数

**文件位置**: `src/routers/router.go`

```go
func (a *API) URLPatterns() []rfb.Route
```

**功能描述**:
- 实现 Go-Chassis 框架的路由接口
- 返回所有路由定义的列表
- 从控制器的 RouteInfo() 自动生成路由

**实现逻辑**:
1. 创建 VideoController 实例
2. 调用控制器的 RouteInfo() 获取路由映射
3. 解析路由映射为标准的 rfb.Route 结构
4. 返回路由列表

**调用关系**:
```
API.URLPatterns()
└── VideoController.RouteInfo()
    └── 返回路由映射: {"/video/*": "GET:GetVideo", "/download/*": "GET:Download", "/test": "GET:Test"}
    └── 通过 parseMethodAndFuncName 解析方法名
```

#### 1.2 parseMethodAndFuncName() 函数

**文件位置**: `src/routers/router.go`

```go
func parseMethodAndFuncName(methodAndFunc string) (string, string)
```

**功能描述**:
- 解析"方法:函数名"格式的字符串
- 返回 HTTP 方法和函数名

**实现逻辑**:
1. 使用 strings.Split 按 ":" 分割字符串
2. 检查分割后数组长度必须为 2
3. 分别返回 HTTP 方法和函数名
4. 格式错误时记录日志并返回空值

**使用示例**:
- 输入: `"GET:GetVideo"` 
- 输出: `("GET", "GetVideo")`

#### 1.3 RegisterRouters() 函数

**文件位置**: `src/routers/beego_router.go`

```go
func RegisterRouters(server https.BeegoServer)
```

**功能描述**:
- 注册所有路由到 Beego 服务器
- 设置全局过滤器
- 注册控制器路由

**实现逻辑**:
1. 全局注册 OverLoadFilter 过滤器（在路由前执行）
2. 调用 registerController 注册 VideoController
3. 支持多种服务器类型（HTTP、HTTPS）

**调用关系**:
```
RegisterRouters(server)
├── server.InsertFilter("*", beego.BeforeRouter, controllers.OverLoadFilter)
└── registerController(server, &controllers.VideoController{})
    └── registerControllerDirectly(server, controller)
        └── controller.RouteInfo() 获取路由映射
        └── server.Router(k, controller, v) 注册路由
        └── registerFilters(server, routeInfo, "") 注册过滤器
```

#### 1.4 registerControllerDirectly() 函数

**文件位置**: `src/routers/beego_router.go`

```go
func registerControllerDirectly(server https.BeegoServer, controller controllers.IController)
```

**功能描述**:
- 直接注册控制器的路由
- 处理路由映射和过滤器

**实现逻辑**:
1. 调用控制器的 RouteInfo() 获取路由信息
2. 遍历 RouteMapping 注册所有路由
3. 记录路由注册日志
4. 调用 registerFilters 注册过滤器

#### 1.5 registerFilters() 函数

**文件位置**: `src/routers/beego_router.go`

```go
func registerFilters(server https.BeegoServer, routeInfo controllers.RouteInfo, routePathPre string)
```

**功能描述**:
- 注册控制器相关的过滤器
- 支持全局和特定路径的过滤器

**实现逻辑**:
1. 注册全局过滤器（匹配所有路由）
2. 如果有路由前缀，注册带前缀的过滤器
3. 支持执行位置控制（BeforeExec 或 AfterExec）

### 2. 控制器函数

#### 2.1 VideoController.RouteInfo() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) RouteInfo() RouteInfo
```

**功能描述**:
- 定义 VideoController 的路由映射
- 返回路由和过滤器信息

**返回的路由映射**:
```go
RouteMapping: map[string]string{
    "/video/*":    "GET:GetVideo",    // 视频流媒体播放
    "/download/*": "GET:Download",   // 文件下载
    "/test":       "GET:Test",        // 测试接口
}
```

#### 2.2 VideoController.Prepare() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) Prepare()
```

**功能描述**:
- 控制器初始化方法
- 在请求处理前调用
- 初始化业务服务依赖

**实现逻辑**:
1. 创建 VideoService 实例
2. 创建 AuthService 实例
3. 为后续请求处理准备服务依赖

#### 2.3 VideoController.GetVideo() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) GetVideo()
```

**功能描述**:
- 处理视频流媒体播放请求
- 支持缓存命中验证
- 设置适当的 HTTP 响应头

**实现逻辑详析**:

1. **性能监控**:
   ```go
   startTime := time.Now()
   // 记录开始时间用于性能分析
   ```

2. **参数提取**:
   ```go
   imei := c.QueryParameter("imei")        // 提取设备IMEI参数
   checkType := c.QueryParameter("checkType") // 提取验证类型参数
   ```

3. **权限验证**:
   ```go
   isValid, err := c.authService.ValidateIMEI(imei, checkType)
   // 验证设备是否有权访问视频内容
   ```

4. **权限处理**:
   ```go
   if err != nil {
       logger.Errorf("[VideoController] IMEI validation failed: %v", err)
       c.Failed(resp.BaseResponse{Code: retcode.InternalFailed, Message: "IMEI validation error"})
       return
   }
   if !isValid {
       logger.Errorf("[VideoController] IMEI validation failed for IMEI: %s", imei)
       c.Failed(resp.BaseResponse{Code: retcode.AuthFailed, Message: "IMEI validation not pass"})
       return
   }
   ```

5. **路径提取**:
   ```go
   videoPath := c.PathParameter(":splat")  // 获取视频文件路径
   ```

6. **获取视频流**:
   ```go
   stream, fileInfo, err := c.videoService.GetVideo(videoPath)
   // 尝试获取视频流和文件信息
   ```

7. **资源清理**:
   ```go
   if stream != nil {
       defer stream.Close()  // 确保流资源正确关闭
   }
   ```

8. **错误处理**:
   ```go
   if err != nil || stream == nil {
       logger.Errorf("get video failed, err: %v", err)
       c.Failed(resp.BaseResponse{Code: 1, Message: err.Error()})
       return
   }
   ```

9. **响应头设置**:
   ```go
   c.FormatHTTPHeader(fileInfo)  // 设置适当的HTTP头
   c.addCorsHeader()            // 添加CORS头
   ```

10. **流式传输**:
    ```go
    var target io.Writer = c.Ctx.ResponseWriter
    if fileInfo.ExtraWriteTarget != nil {
        target = io.MultiWriter(c.Ctx.ResponseWriter, fileInfo.ExtraWriteTarget)
    }
    // 支持多目标写入（如日志记录）
    ```

11. **数据传输**:
    ```go
    if written, err := io.Copy(target, stream); err != nil {
        logger.Errorf("[VideoController] write file[%s] failed, err: %v, hasWritten: %d", videoPath, err, written)
        c.Failed(resp.BaseResponse{Code: 1, Message: err.Error()})
        return
    }
    ```

12. **性能日志**:
    ```go
    logger.Infof("[VideoController] file[%s] use time: %v", videoPath, time.Since(startTime))
    ```

#### 2.4 VideoController.FormatHTTPHeader() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) FormatHTTPHeader(fileInfo *storage.FileInfo)
```

**功能描述**:
- 为视频流设置适当的 HTTP 响应头
- 优化缓存和流媒体播放

**设置的头信息**:
- `Last-Modified`: 文件修改时间
- `ETag`: 文件内容的哈希值
- `content-md5`: 文件 MD5 哈希
- `Content-Length`: 文件大小
- `Accept-Ranges`: 支持范围请求
- `Connection`: 保持连接
- `cache-control`: 缓存控制策略
- `Content-Type`: 根据文件扩展名设置媒体类型

#### 2.5 VideoController.addCorsHeader() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) addCorsHeader()
```

**功能描述**:
- 设置跨域资源共享 (CORS) 头信息
- 允许前端应用跨域访问视频API

**设置的 CORS 头**:
- `Access-Control-Allow-Origin`: 允许所有来源
- `Access-Control-Allow-Methods`: 允许 GET 和 OPTIONS 方法
- `Access-Control-Allow-Headers`: 允许 Range 等必要的请求头
- `Access-Control-Expose-Headers`: 暴露响应头给前端

#### 2.6 VideoController.Download() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) Download()
```

**功能描述**:
- 处理文件下载请求
- 模拟 Muen 媒体服务器的下载功能
- 设置文件下载相关的 HTTP 头

**实现逻辑**:

1. **路径处理**:
   ```go
   videoPath := c.PathParameter(":splat")  // 获取下载文件路径
   logger.Infof("video path %s", videoPath) // 记录下载路径
   ```

2. **文件获取**:
   ```go
   rc, size, err := c.videoService.Download(videoPath)
   // 获取文件读取器和文件大小
   ```

3. **错误处理**:
   ```go
   if err != nil {
       logger.Errorf("Download failed, err is [%v]", err)
       c.Failed(resp.BaseResponse{Code: 1, Message: err.Error()})
       return
   }
   ```

4. **资源清理**:
   ```go
   defer func() {
       if err = rc.Close(); err != nil {
           logger.Errorf("failed close file %s, err: %v", videoPath, err)
       }
   }()
   ```

5. **下载头设置**:
   ```go
   c.AddHeader("Content-Disposition", "attachment; filename="+filepath.Base(videoPath))
   c.AddHeader("Content-Length", fmt.Sprintf("%d", size))
   ```

6. **文件传输**:
   ```go
   if _, err = io.Copy(c.ResponseWriter(), rc); err != nil {
       logger.Errorf("Download copy file to client failed, err is [%v]", err)
       c.Failed(resp.BaseResponse{Code: 1, Message: err.Error()})
       return
   }
   ```

7. **成功响应**:
   ```go
   response := resp.DataResponse{
       BaseResponse: resp.BaseResponse{
           Code:    200,
           Message: "download success",
       },
       Data: true,
   }
   c.OK(response)  // 返回下载成功响应
   ```

#### 2.7 VideoController.Test() 函数

**文件位置**: `src/controllers/video_controller.go`

```go
func (c *VideoController) Test()
```

**功能描述**:
- 简单的测试接口
- 用于服务健康检查和基本功能验证

**实现逻辑**:
1. 返回固定的成功响应
2. 返回 `{code: 0, message: "test success", data: true}`

### 3. 服务层函数

#### 3.1 AuthServiceImpl.ValidateIMEI() 函数

**文件位置**: `src/service/auth_service.go`

```go
func (s *AuthServiceImpl) ValidateIMEI(imei string, checkType string) (bool, error)
```

**功能描述**:
- 验证设备 IMEI 的合法性
- 调用远程服务进行验证
- 记录验证过程和结果

**实现逻辑**:

1. **开始验证**:
   ```go
   startTime := time.Now()
   logger.Infof("[AuthServiceImpl] Start to validate IMEI")
   ```

2. **远程验证**:
   ```go
   isValid, err := s.remote.PostValidateIMEI(imei, checkType)
   // 调用远程认证服务
   ```

3. **错误处理**:
   ```go
   if err != nil {
       logger.Errorf("[AuthServiceImpl] IMEI validation failed")
       return false, fmt.Errorf("IMEI validation failed: %v", err)
   }
   ```

4. **日志记录**:
   ```go
   logger.Infof("[AuthServiceImpl] IMEI validation completed, result: %v, cost: %v",
       isValid, time.Since(startTime))
   return isValid, nil
   ```

#### 3.2 VideoServiceImpl.GetVideo() 函数

**文件位置**: `src/service/VideoService.go`

```go
func (s *VideoServiceImpl) GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error)
```

**功能描述**:
- 获取视频流和文件信息
- 实现本地缓存优先策略
- 集成远程服务获取和告警管理

**实现逻辑**:

1. **性能记录**:
   ```go
   startTime := time.Now()
   videoPath = filepath.Clean(videoPath)  // 清理路径
   ```

2. **缓存检查**:
   ```go
   if s.storage.Exist(videoPath) {
       if data, localFileInfo, err := s.storage.Get(videoPath); err == nil {
           logger.Infof("[VideoServiceImpl] hit video[%s] success, cost: %v", videoPath, time.Since(startTime))
           return data, localFileInfo, nil
       }
   }
   // 缓存命中成功立即返回
   ```

3. **远程获取**:
   ```go
   data, fileInfoFromRemote, err := s.remote.GetVideo(videoPath)
   // 尝试从远程服务器获取
   ```

4. **错误处理和告警**:
   ```go
   if err != nil {
       logger.Errorf("[VideoServiceImpl] failed to get video[%s] from MUEN, err: %s", videoPath, err)
       s.alarm.SendAlarm(AlarmId300020, "Failed to get video content"+videoPath)
       return nil, nil, err
   }
   ```

5. **清除告警**:
   ```go
   s.alarm.ClearAlarm(AlarmId300020, "Succeed to get video content")
   ```

6. **缓存保存**:
   ```go
   localFileInfo, err := s.storage.Cache(videoPath)
   // 将文件保存到本地缓存
   ```

7. **缓存信息更新**:
   ```go
   localFileInfo.Hash = fileInfoFromRemote.Hash
   localFileInfo.AddFinalizer(fileInfoFromRemote.Finalizer)
   // 更新哈希值和清理函数
   ```

#### 3.3 VideoServiceImpl.Download() 函数

**文件位置**: `src/service/VideoService.go`

```go
func (s *VideoServiceImpl) Download(videoPath string) (io.ReadCloser, int64, error)
```

**功能描述**:
- 文件下载功能
- 直接读取本地文件系统
- 返回文件读取器和文件大小

**实现逻辑**:

1. **文件检查**:
   ```go
   fi, err := os.Stat("/" + videoPath)
   if os.IsNotExist(err) {
       return nil, 0, fmt.Errorf("failed to get file[%s], err: %v", "/"+videoPath, err)
   }
   ```

2. **文件打开**:
   ```go
   file, err := os.Open(videoPath)
   if err != nil {
       return nil, 0, fmt.Errorf("failed to open file[%s], err: %v", "/"+videoPath, err)
   }
   // 返回文件读取器和文件大小
   return file, fi.Size(), nil
   ```

## 调用关系图

### 1. 整体架构调用关系

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP 请求                                │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                     VideoController                         │
│  ┌────────────────────┐  ┌────────────────────┐            │
│  │  GetVideo()        │  │  Download()        │            │
│  │  Download()        │  │  Test()            │            │
│  └────────────────────┘  └────────────────────┘            │
└─────────────────────┬───────────────────────────────────────┘
                      │
              ┌───────┴───────┐
              ▼               ▼
┌─────────────────────┐ ┌─────────────────────┐
│    AuthService      │ │   VideoService      │
│                     │ │                     │
│ ┌──────────────────┐ │ │ ┌──────────────────┐ │
│ │ ValidateIMEI()   │ │ │ │ GetVideo()       │ │
│ └──────────────────┘ │ │ │ Download()       │ │
│                     │ │ └──────────────────┘ │
└─────────────────────┘ │                     │
              ┌───────┐ ▼ ┌──────────────────┐ │
              │       │   │ ┌──────────────────┐ │
              ▼       │   │ │ Storage 接口     │ │
┌─────────────────────┐ │   │ │                 │ │
│    Remote Service   │ │   │ │ Cache()         │ │
│                     │ │   │ │ Get()           │ │
│ ┌──────────────────┐ │   │ │ Exist()         │ │
│ │ PostValidateIMEI │ │   │ └──────────────────┘ │
│ │ GetVideo()       │ │   │                     │
│ └──────────────────┘ │   └─────────────────────┘
└─────────────────────┘                │
                                     ┌──┴───┐
                                     ▼      │
                              ┌────────────────┐
                              │   本地存储      │
                              │ (LocalStorage)  │
                              └────────────────┘
```

### 2. 详细函数调用关系

#### 2.1 GetVideo() 调用链

```
HTTP GET /video/xxx → VideoController.GetVideo()
├── VideoController.Prepare()
│   ├── service.NewVideoService()
│   └── service.NewAuthService()
├── c.authService.ValidateIMEI()
│   ├── s.remote.PostValidateIMEI() (远程认证)
│   └── 记录认证结果和耗时
├── c.videoService.GetVideo()
│   ├── s.storage.Exist() (检查缓存)
│   ├── s.storage.Get() (缓存命中时返回)
│   ├── s.remote.GetVideo() (远程获取)
│   ├── s.alarm.SendAlarm() (失败时发送告警)
│   ├── s.alarm.ClearAlarm() (成功时清除告警)
│   └── s.storage.Cache() (保存到缓存)
├── c.FormatHTTPHeader() (设置HTTP头)
├── c.addCorsHeader() (设置CORS头)
├── io.Copy() (数据传输)
└── 记录处理耗时
```

#### 2.2 Download() 调用链

```
HTTP GET /download/xxx → VideoController.Download()
├── VideoController.Prepare()
│   ├── service.NewVideoService()
│   └── service.NewAuthService()
├── c.videoService.Download()
│   ├── os.Stat() (检查文件存在性)
│   └── os.Open() (打开文件)
├── c.WriteHeaderAndJSON() (响应设置)
└── io.Copy() (文件传输)
```

#### 2.3 路由注册调用链

```
main() → startInternalServer() / startExternalServer()
├── https.NewHttpServer() / https.NewHttpsServer()
├── routers.RegisterRouters()
│   ├── server.InsertFilter() (注册全局过滤器)
│   └── registerController()
│       ├── registerControllerDirectly()
│       │   ├── controller.RouteInfo() (获取路由映射)
│       │   ├── server.Router() (注册单个路由)
│       │   └── registerFilters() (注册过滤器)
│       └── registerFilters()
└── server.Run() (启动服务器)
```

### 3. 服务初始化调用关系

```
main()
├── conf.Instance() & flagutil.Parse() (配置初始化)
├── tasks.InitCronTasks() (定时任务初始化)
├── initGSF() (华为框架初始化)
├── ntp.Init() (时间同步初始化)
├── runLogInit() (日志初始化)
├── remote.InitMuenClient() (远程客户端初始化)
├── service.CleanAllActiveAlarm() (告警清理)
└── server 启动和阻塞
```

## HTTP API接口

### 1. 视频流媒体接口

#### GET /video/*

**功能**: 视频流媒体播放

**请求参数**:
```
imei: 设备 IMEI 号 (必需)
checkType: 验证类型 (必需)
```

**响应头**:
```
Last-Modified: Tue, 15 Nov 2024 12:00:00 GMT
ETag: "abc123"
content-md5: abc123
Content-Length: 1048576
Accept-Ranges: bytes
Connection: keep-alive
cache-control: public,max-age=31536000
Content-Type: video/mp4 (或 video/mp2t, application/vnd.apple.mpegurl)
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: false
Access-Control-Allow-Methods: GET, OPTIONS
Access-Control-Allow-Headers: Range, Origin, Content-Type
Access-Control-Expose-Headers: Content-Range, Last-Modified, Etag,content-md5, Content-Length, Accept-Ranges, Vary
```

**响应状态码**:
- 200: 成功返回视频流
- 400: IMEI 验证失败
- 500: 服务器内部错误

**调用示例**:
```
GET /video/sample.mp4?imei=123456789012345&checkType=device
```

### 2. 文件下载接口

#### GET /download/*

**功能**: 文件下载

**请求参数**: 无 (路径参数包含文件路径)

**响应头**:
```
Content-Disposition: attachment; filename=filename.mp4
Content-Length: 1048576
```

**响应体**: 文件内容

**调用示例**:
```
GET /download/sample.mp4
```

### 3. 健康检查接口

#### GET /test

**功能**: 服务健康检查和基本功能验证

**响应示例**:
```json
{
  "code": 0,
  "msg": "test success",
  "data": true
}
```

## 路由注册机制

### 1. 双框架支持

MediaCacheService 支持两种路由注册方式：

#### Beego 框架路由

**注册机制**:
- 通过 `routers.RegisterRouters()` 函数
- 使用 `server.Router()` 方法注册具体路由
- 支持控制器方法的直接映射

**特点**:
- 简单直接的路由注册
- 适合中小型微服务
- 支持灵活的中间件配置

#### Go-Chassis 框架路由

**注册机制**:
- 通过 `API.URLPatterns()` 方法
- 返回路由定义列表
- 适合微服务治理

**特点**:
- 支持服务发现和治理
- 提供更完整的企业级功能
- 适合分布式系统

### 2. 路由配置方式

#### 控制器端路由定义

```go
func (c *VideoController) RouteInfo() RouteInfo {
    return RouteInfo{
        RouteMapping: map[string]string{
            "/video/*":    "GET:GetVideo",    // 视频流媒体
            "/download/*": "GET:Download",   // 文件下载
            "/test":       "GET:Test",        // 测试接口
        },
    }
}
```

#### 路由注册流程

1. **初始化阶段**:
   ```go
   // main.go
   server := https.NewHttpServer(ip, port)
   routers.RegisterRouters(server)
   ```

2. **路由注册**:
   ```go
   // beego_router.go
   func RegisterRouters(server https.BeegoServer) {
       server.InsertFilter("*", beego.BeforeRouter, controllers.OverLoadFilter)
       registerController(server, &controllers.VideoController{})
   }
   ```

3. **控制器注册**:
   ```go
   // beego_router.go
   func registerControllerDirectly(server https.BeegoServer, controller controllers.IController) {
       routeInfo := controller.RouteInfo()
       for k, v := range routeInfo.RouteMapping {
           server.Router(k, controller, v)  // 注册路由
       }
   }
   ```

## 中间件机制

### 1. 过滤器类型

#### FilterAction 枚举

**定义位置**: `src/controllers/controller.go`

```go
type FilterAction string

const (
    Before FilterAction = "before"  // 在控制器方法执行前
    After  FilterAction = "after"   // 在控制器方法执行后
)
```

#### 自定义过滤器

**实现方式**:
```go
type RouteInfo struct {
    RouteMapping map[string]string
    Filters      map[FilterAction]beego.FilterFunc
}
```

### 2. 负载过滤器

#### OverLoadFilter 实现

**定义位置**: `src/controllers/filter.go`

**功能**:
- 防止系统过载
- 返回 HTTP 429 状态码
- 支持自定义过载检测逻辑

**注册方式**:
```go
server.InsertFilter("*", beego.BeforeRouter, controllers.OverLoadFilter)
```

**执行时机**: 在所有路由处理之前全局执行

### 3. CORS 过滤器

#### addCorsHeader 实现位置**: `src/controllers/video_controller.go`

**功能**:
- 设置跨域资源共享头
- 支持多种请求方法和头
- 适应流媒体播放需求

**设置的头**:
```go
c.AddHeader("Access-Control-Allow-Origin", "*")
c.AddHeader("Access-Control-Allow-Methods", "GET, OPTIONS")
c.AddHeader("Access-Control-Allow-Headers", "Range, Origin, Content-Type")
c.AddHeader("Access-Control-Expose-Headers", "Content-Range, Last-Modified, Etag,content-md5, Content-Length, Accept-Ranges, Vary")
```

### 4. 过滤器注册流程

```go
func registerFilters(server https.BeegoServer, routeInfo controllers.RouteInfo, routePathPre string) {
    // 注册全局过滤器（匹配所有路由）
    for k, v := range routeInfo.Filters {
        var pos = beego.BeforeExec
        if k == controllers.After {
            pos = beego.AfterExec
        }
        server.InsertFilter("/*", pos, v, beego.WithReturnOnOutput(false))
    }

    // 注册带前缀的过滤器（匹配特定前缀下的所有子路由）
    if routePathPre != "" {
        for k, v := range routeInfo.Filters {
            var pos = beego.BeforeExec
            if k == controllers.After {
                pos = beego.AfterExec
            }
            server.InsertFilter(routePathPre+"/*", pos, v, beego.WithReturnOnOutput(false))
        }
    }
}
```

## 总结

MediaCacheService 的路由系统具有以下特点：

1. **架构清晰**: 采用控制器、服务、存储三层架构，职责明确
2. **多框架支持**: 同时支持 Beego 和 Go-Chassis 框架
3. **企业级特性**: 集成华为 CSP 平台的各种企业功能
4. **性能优化**: 实现本地缓存、流媒体传输、CORS 支持
5. **安全性**:完整的认证机制和权限控制
6. **可扩展性**: 接口化设计，便于功能扩展

该路由系统设计合理，功能完整，适合企业级媒体缓存服务的需求。