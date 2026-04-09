# MediaCacheService - util 目录分析文档

> 生成时间: 2026-03-24
> 路径: D:\CloudCellular\MediaCacheService\src\util

---

## 📁 目录结构

```
util/
├── flag/
│   ├── flags.go           # 命令行参数解析工具
│   └── flags_test.go      # flag 解析测试文件
├── response_util.go        # HTTP 响应标准化工具
└── sys/
    └── sys.go            # 系统操作接口
```

---

## 📊 总体概览

util 目录包含 3 个主要模块,提供核心基础设施支持:

| 模块 | 包名 | 功能描述 | 文件数 |
|-----|------|---------|--------|
| flag | flagutil | 命令行参数解析和结构体标签处理 | 2 |
| response | response | HTTP 响应标准化封装 | 1 |
| sys | sys | 系统级操作(目录大小、文件管理) | 1 |

---

## 🔧 接口定义

### 1. Interface (系统操作接口)

**位置**: `D:\CloudCellular\MediaCacheService\src\util\sys\sys.go`

```go
type Interface interface {
    // SysDirSize 递归判断目录的大小, 单位MB
    SysDirSize(dirPath string) (int, error)
    DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
}
```

**功能描述**:
- 提供系统级文件系统管理操作
- 支持目录大小计算(以MB为单位)
- 支持基于访问时间的文件清理功能

**实现结构体**: `sysImpl`

---

## 🏗️ 结构体定义

### 1. sysImpl

**位置**: `D:\CloudCellular\MediaCacheService\src\util\sys\sys.go` (第31-32行)

```go
type sysImpl struct {
}
```

**功能描述**:
- `Interface` 接口的默认实现
- 使用系统命令(`du`, `find`)执行文件系统操作
- 包含错误处理和日志记录功能

**接口符合性**: `var _ Interface = &sysImpl{}`

---

## 📦 函数详解

### response_util.go - 响应工具模块

#### Success() - 成功响应构造器

**函数签名**:
```go
func Success(data interface{}) resp.DataResponse
```

**位置**: `D:\CloudCellular\MediaCacheService\src\util\response_util.go` (第7-14行)

**功能描述**:
创建标准化的 HTTP 成功响应,包含状态码200和指定的数据负载

**返回值**:
- 返回 `resp.DataResponse` 结构体实例
- 包含 `BaseResponse` 字段: `Code: 200`, `Message: "success"`
- 包含 `Data` 字段: 用户传入的数据

**调用关系**:
- 无内部调用
- 被 controllers 层广泛使用

**代码分析**:
```go
func Success(data interface{}) resp.DataResponse {
    return resp.DataResponse{
        BaseResponse: resp.BaseResponse{
            Code:    200,
            Message: "success",
        },
        Data: data,
    }
}
```

---

### flag/flags.go - 命令行参数解析模块

#### Parse() - 结构体标签解析主函数

**函数签名**:
```go
func Parse(obj interface{}) interface{}
```

**位置**: `D:\CloudCellular\MediaCacheService\src\util\flag\flags.go` (第13-31行)

**功能描述**:
递归解析结构体的 `flag` tag,并将结构体字段注册为命令行参数

**参数**:
- `obj interface{}` - 传入的结构体指针(包含默认值)

**返回值**:
- `interface{}` - 解析后的新对象值

**实现机制**:
1. 使用反射获取结构体值
2. 递归处理嵌套结构体
3. 根据字段类型注册对应的 flag
4. 调用 `flag.Parse()` 解析命令行参数
5. 将解析后的值写回原始对象

**调用关系**:
- 调用 `parseStruct()` (内部辅助函数)
- 调用 `flag.Parse()` (标准库)
- 被 main.go 启动时调用

**代码详细分析**:
```go
func Parse(obj interface{}) interface{} {
    // 1. 获取反射值
    v := reflect.ValueOf(obj)

    // 2. 处理指针类型,解引用到实际结构体
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    // 3. 验证必须是结构体
    if v.Kind() != reflect.Struct {
        return nil
    }

    // 4. 创建同类型的空对象用于接收解析后的值
    newObj := reflect.New(v.Type()).Elem()

    // 5. 递归解析结构体字段
    parseStruct(v, newObj, "")

    // 6. 执行标准 flag 解析
    flag.Parse()

    // 7. 将解析后的新对象值写回原始对象
    v.Set(newObj)

    // 8. 返回新对象
    return newObj
}
```

**支持的数据类型**:
- `string`
- `int`, `int64`
- `uint`, `uint64`
- `bool`
- `float64`

---

#### parseStruct() - 递归结构体解析

**函数签名**:
```go
func parseStruct(defaultObj, newObj reflect.Value, prefix string)
```

**位置**: `D:\CloudCellular\MediaCacheService\src\util\flag\flags.go` (第34-87行)

**权限**: 包私有函数(小写开头)

**功能描述**:
递归解析结构体字段并注册为命令行 flag,处理嵌套结构体

**参数**:
- `defaultObj reflect.Value` - 包含默认值的原始结构体
- `newObj reflect.Value` - 待解析的新结构体(可修改)
- `prefix string` - 嵌套结构体的前缀标识(如 "parent-child")

**调用关系**:
- 被 `Parse()` 调用
- 递归调用自身处理嵌套结构体

**实现细节**:

1. **遍历所有字段**:
```go
for i := 0; i < newObj.NumField(); i++ {
    field := newObj.Type().Field(i)
    fieldValue := newObj.Field(i)
    defaultFieldValue := defaultObj.Field(i)
```

2. **处理 flag tag**:
```go
flagName := field.Tag.Get("flag")
fullFlagName := flagName

// 根据 prefix 和 flagName 组合完整 flag 名称
if flagName == "" {
    fullFlagName = prefix
} else if prefix == "" {
    fullFlagName = flagName
} else {
    fullFlagName = prefix + "-" + flagName  // 嵌套结构体: parent-child-field
}
```

3. **递归处理嵌套结构体**:
```go
if fieldValue.Kind() == reflect.Struct {
    parseStruct(defaultFieldValue, fieldValue, fullFlagName)
    continue
}
```

4. **跳过无 flag tag 的非结构体字段**:
```go
if flagName == "" {
    continue
}
```

5. **创建可寻址指针**:
```go
if !fieldValue.CanAddr() {
    continue
}
fieldPtr := reflect.NewAt(field.Type, fieldValue.Addr().UnsafePointer()).Interface()
```

6. **根据字段类型注册 flag**:
```go
switch field.Type.Kind() {
case reflect.String:
    flag.StringVar(fieldPtr.(*string), fullFlagName, defaultFieldValue.String(), desc)
case reflect.Int, reflect.Int64:
    flag.IntVar(fieldPtr.(*int), fullFlagName, int(defaultFieldValue.Int()), desc)
case reflect.Uint, reflect.Uint64:
    flag.UintVar(fieldPtr.(*uint), fullFlagName, uint(defaultFieldValue.Int()), desc)
case reflect.Bool:
    flag.BoolVar(fieldPtr.(*bool), fullFlagName, defaultFieldValue.Bool(), desc)
case reflect.Float64:
    flag.Float64(fieldPtr.(*float64), fullFlagName, defaultFieldValue.Float(), desc)
default:
    panic(fmt.Sprintf("不支持的类型: %s", field.Type))
}
```

**错误处理**:
- 遇到不支持的类型会 panic 并报错

---

### sys/sys.go - 系统操作模块

#### New() - 系统操作工厂函数

**函数签名**:
```go
func New() Interface
```

**位置**: `D:\CloudCellular\MediaCacheService\src\util\sys\sys.go` (第25-27行)

**功能描述**:
工厂函数,返回系统操作接口的默认实现

**返回值**:
- `Interface` - 返回 `sysImpl` 实例

**调用关系**:
- 被 service 层调用获取系统操作实例

**代码分析**:
```go
func New() Interface {
    return &sysImpl{}
}
```

---

#### DeleteInactiveFile() - 删除不活跃文件

**函数签名**:
```go
func (s sysImpl) DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
```

**位置**: `D:\CloudCellular\MediaCacheService\src\util\sys\sys.go` (第34-54行)

**接收者**: `s sysImpl`

**功能描述**:
删除指定目录中超过一定天数未访问的文件

**参数**:
- `dirPath string` - 目标目录路径
- `threshold int` - 访问时间阈值(天),超过此天数的文件将被删除
- `timeout time.Duration` - 命令执行超时时间,防止命令挂起

**返回值**:
- `error` - 操作错误(成功时为 nil)

**实现机制**:

1. **创建带超时的 context**:
```go
ctx, cancel := goctx.WithTimeout(goctx.Background(), timeout)
defer cancel()
```

2. **构造 find 命令**:
```go
cmd := exec.CommandContext(ctx, "find", dirPath,
    "-type", "f",           // 只匹配普通文件
    "-atime", fmt.Sprintf("+%d", threshold),  // 访问时间超过 threshold 天
    "-delete")              // 删除匹配的文件
```

3. **执行命令并获取输出/错误**:
```go
output, err := cmd.CombinedOutput()
```

4. **错误处理和日志**:
```go
if err != nil {
    logger.Errorf("exec 'find and delete file' command failed: %v, command output %v", err, string(output))
    return err
}

logger.Infof("success to exec 'find and delete file' command, dirpath: %v, threshold: %v, timeout: %v, output: %v",
    dirPath, threshold, timeout, string(output))
```

**系统命令说明**:
- `find <dir> -type f -atime +<n> -delete`
- `-type f`: 仅匹配普通文件(不包括目录、符号链接等)
- `-atime +<n>`: 文件最后访问时间超过 n 天
- `-delete`: 直接删除匹配到的文件

**调用关系**:
- 被 storage 层的清理任务调用
- 调用 `exec.CommandContext()` (标准库)
- 调用 `logger.*()` (日志模块)

**使用场景**:
- 定期清理长时间未访问的缓存文件
- 释放存储空间
- 基于文件访问时间的智能清理策略

---

#### SysDirSize() - 目录大小计算

**函数签名**:
```go
func (s sysImpl) SysDirSize(dirPath string) (int, error)
```

**位置**: `D:\CloudCellular\MediaCacheService\src\util\sys\sys.go` (第56-77行)

**接收者**: `s sysImpl`

**功能描述**:
递归计算指定目录的总大小(单位:MB)

**参数**:
- `dirPath string` - 目标目录路径

**返回值**:
- `int` - 目录大小(MB)
- `error` - 操作错误(成功时为 nil)

**实现机制**:

1. **构造 du 命令**:
```go
cmd := exec.Command("du", "-sm", dirPath)
```

2. **执行命令并获取输出**:
```go
output, err := cmd.Output()
```

3. **错误处理**:
```go
if err != nil {
    return 0, fmt.Errorf("failed to execute 'du' command: %w", err)
}
```

4. **解析输出**:
```go
// du 输出格式: "12345 ./testdir"
parts := strings.Fields(string(output))

if len(parts) == 0 {
    return 0, fmt.Errorf("invalid output format from 'du' command: empty result")
}

sizeMb, err := strconv.ParseInt(parts[0], 10, 64)
if err != nil {
    return 0, fmt.Errorf("failed to parse size from 'du' output: %w", err)
}
```

5. **日志记录**:
```go
logger.Infof("success to exec 'du' command. dirpath: %v, sizeMb: %v", dirPath, sizeMb)
```

6. **返回结果**:
```go
return int(sizeMb), nil
```

**系统命令说明**:
- `du -sm <dir>`
- `-s`: 仅输出总大小(不显示子目录)
- `-m`: 以 MB 为单位(1块=1MB)

**输出格式示例**:
```
12345 ./testdir
```
- `12345`: 大小(MB)
- `./testdir`: 目录路径

**调用关系**:
- 被 storage 层调用监控存储使用情况
- 被 service 层调用显示缓存统计信息
- 调用 `exec.Command()` (标准库)
- 调用 `logger.*()` (日志模块)

**错误处理**:
- 空输出: 返回错误
- 解析失败: 返回错误
- 命令执行失败: 返回错误

---

## 🔗 函数调用关系图

### 模块级调用关系

```
┌─────────────────────────────────────────────────────────────┐
│  Util Package Call Graph                                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  response_util.go                                           │
│  └── Success() ──> 无内部调用,被外部使用                     │
│                                                              │
│  flag/flags.go                                              │
│  ├── Parse() ──> parseStruct() ──> [递归]                    │
│  │               └─> flag.Parse() (标准库)                   │
│  └── parseStruct() ──> [自身递归处理嵌套结构体]              │
│                     └─> flag.*Var() (根据类型注册)          │
│                                                              │
│  sys/sys.go                                                 │
│  ├── New() ──> 创建 sysImpl 实例                           │
│  ├── DeleteInactiveFile() ──> exec.CommandContext()         │
│  │                           ├─> logger.Errorf()            │
│  │                           └─> logger.Infof()             │
│  └── SysDirSize() ──> exec.Command()                        │
│                      ├─> logger.Infof()                     │
│                      └─> strconv.ParseInt()                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 跨模块调用关系

```
┌─────────────────────────────────────────────────────────────┐
│  External Callers                                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ✅ response.Success()                                      │
│     ├─> controllers/* 层                                    │
│     └─> API 处理器中返回成功响应                             │
│                                                              │
│  ✅ flagutil.Parse()                                        │
│     ├─> main.go (应用启动时)                                │
│     └─> 解析命令行配置参数                                   │
│                                                              │
│  ✅ sys.New()                                               │
│     ├─> service/* 层                                         │
│     └─> storage/* 层                                         │
│                                                              │
│  ✅ sys.SysDirSize()                                        │
│     ├─> storage.* 缓存统计                                  │
│     └─> service.* 显示存储使用情况                           │
│                                                              │
│  ✅ sys.DeleteInactiveFile()                                │
│     ├─> storage.* 清理任务                                   │
│     └─> tasks.* 后台清理作业                                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧩 关键技术点

### 反射 (Reflection)
- **位置**: `flag/flags.go`
- **用途**: 动态解析结构体字段和标签,自动注册命令行参数
- **优势**: 减少重复代码,支持复杂嵌套结构体
- **性能**: 仅在应用启动时执行,对运行时性能无影响

### 反射类型检查
- **支持类型**: string, int/int64, uint/uint64, bool, float64
- **类型验证**: 运行时检查字段类型,不支持的类型会 panic
- **指针转换**: 通过 `reflect.NewAt()` 安全转换字段指针

### 系统命令执行
- **位置**: `sys/sys.go`
- **命令**:
  - `du -sm`: 计算目录大小
  - `find -type f -atime +n -delete`: 删除旧文件
- **超时控制**: 使用 `context.WithTimeout()` 防止命令挂起
- **错误处理**: 捕获命令输出和错误码

### 标准库集成
- **flag package**: 标准命令行参数解析
- **exec package**: 外部命令执行
- **context package**: 超时控制和上下文管理
- **reflect package**: 运行时反射操作

---

## 📝 使用示例

### 1. 使用 flagutil 解析命令行参数

```go
package main

import (
    "flagutil"
    "fmt"
)

type Config struct {
    Port    int    `flag:"port" desc:"服务端口"`
    Debug   bool   `flag:"debug" desc:"调试模式"`
    LogFile string `flag:"log-file" desc:"日志文件路径"`

    Database struct {
        Host     string `flag:"db-host" desc:"数据库主机"`
        Port     int    `flag:"db-port" desc:"数据库端口"`
        User     string `flag:"db-user" desc:"数据库用户"`
        Password string `flag:"db-password" desc:"数据库密码"`
    } `flag:"database" desc:"数据库配置"`
}

func main() {
    // 设置默认值
    config := &Config{
        Port:  8080,
        Debug: false,
        Database: struct {
            Host     string
            Port     int
            User     string
            Password string
        }{
            Host: "localhost",
            Port: 3306,
        },
    }

    // 解析命令行参数
    flagutil.Parse(config)

    // 使用解析后的配置
    fmt.Printf("Port: %d\n", config.Port)
    fmt.Printf("DB Host: %s\n", config.Database.Host)
}
```

**命令行使用**:
```bash
./app --port=9000 --debug=true --database-db-host=192.168.1.1 --database-db-port=5432
```

---

### 2. 使用 response.Success() 返回成功响应

```go
package handlers

import (
    "response"
    "net/http"
)

type UserInfo struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func GetUser(w http.ResponseWriter, r *http.Request) {
    // ... 查询用户逻辑 ...

    user := UserInfo{
        ID:    "123",
        Name:  "张三",
        Email: "zhangsan@example.com",
    }

    // 返回成功响应
    resp := response.Success(user)
    json.NewEncoder(w).Encode(resp)
}
```

**响应格式**:
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "123",
        "name": "张三",
        "email": "zhangsan@example.com"
    }
}
```

---

### 3. 使用 sys 包进行系统操作

```go
package storage

import (
    "sys"
    "time"
)

type CacheManager struct {
    sysOp sys.Interface
}

func NewCacheManager() *CacheManager {
    return &CacheManager{
        sysOp: sys.New(),
    }
}

// 获取缓存目录大小
func (cm *CacheManager) GetCacheSize() (int, error) {
    return cm.sysOp.SysDirSize("/var/cache/media")
}

// 清理 30 天未访问的文件,超时 1 分钟
func (cm *CacheManager) CleanOldFiles() error {
    return cm.sysOp.DeleteInactiveFile(
        "/var/cache/media",
        30,                     // 30天阈值
        time.Minute,           // 1分钟超时
    )
}
```

---

## 🔍 设计模式分析

### 1. 工厂模式 (Factory Pattern)

**位置**: `sys/sys.go`

```go
var NewFunc = New

func New() Interface {
    return &sysImpl{}
}
```

**特点**:
- 提供 `New()` 工厂函数创建实例
- 返回接口类型而非具体实现类型
- 支持依赖注入和单元测试

---

### 2. 接口隔离原则 (Interface Segregation)

```go
type Interface interface {
    SysDirSize(dirPath string) (int, error)
    DeleteInactiveFile(dirPath string, threshold int, timeout time.Duration) error
}
```

**特点**:
- 定义最小接口,仅包含必要方法
- 依赖接口而非具体实现
- 便于测试和扩展

---

### 3. 递归处理模式

**位置**: `flag/flags.go`

```go
func parseStruct(defaultObj, newObj reflect.Value, prefix string) {
    // 遍历字段
    for i := 0; i < newObj.NumField(); i++ {
        // 处理嵌套结构体
        if fieldValue.Kind() == reflect.Struct {
            parseStruct(defaultFieldValue, fieldValue, fullFlagName)
            continue
        }
        // ...
    }
}
```

**特点**:
- 递归处理支持任意深度的嵌套结构体
- 通过 prefix 参数传递层次结构信息
- 简化复杂嵌套数据处理

---

## 🛠️ 扩展建议

### 1. flagutil 扩展

**当前限制**:
- 仅支持基本数据类型
- 不支持切片、数组、映射等复合类型

**扩展方向**:
```go
// 支持字符串切片
[]string `flag:"tags" desc:"标签列表"`

// 支持时间类型
time.Time `flag:"created" desc:"创建时间"`
```

---

### 2. response util 扩展

**当前功能**:
- 仅提供成功响应

**建议扩展**:
```go
// 错误响应
func Error(code int, message string) resp.DataResponse

// 分页响应
func Paginated(data interface{}, page, pageSize, total int) resp.PaginatedResponse

// 文件响应
func File(filePath string) http.HandlerFunc
```

---

### 3. sys util 扩展

**当前功能**:
- 目录大小计算
- 删除不活跃文件

**建议扩展**:
```go
// 磁盘空间检查
type DiskInfo struct {
    Total      int64
    Used       int64
    Available  int64
    UsagePercent float64
}
GetDiskInfo(path string) (DiskInfo, error)

// 文件统计
type FileStats struct {
    FileCount  int
    DirCount   int
    TotalSize  int64
}
GetFileStats(dirPath string) (FileStats, error)

// 压缩目录
CompressDir(src, dst string) error

// 跨平台支持
// 当前仅支持 Unix-like 系统(find, du 命令)
// 可添加 Windows 支持(dir, forfiles 命令)
```

---

## 📊 性能考虑

### flagutil 性能
- **执行时机**: 应用启动时(一次性)
- **性能影响**: 可忽略不计(仅反射解析和 flag 注册)
- **反射开销**: 仅启动时,无运行时影响

### sys 性能
- **命令执行**: 每次 API 调用都会执行系统命令
- **建议场景**:
  - 不频繁操作(如每日清理任务)
  - 监控指标(如每10分钟检查一次)
- **避免**: 高频调用会因命令执行而影响性能

### response 性能
- **无状态函数**: 纯数据构造,无性能瓶颈
- **适用场景**: 高并发 API 响应

---

## 🧪 测试覆盖

### flag/flags_test.go

提供完整的单元测试,覆盖:
- 基本类型测试: string, int, uint, bool, float64
- 嵌套结构体测试
- 默认值测试
- Flag 名称组合测试

**测试模式**: 测试表(table-driven test)

---

## 🔒 错误处理策略

### sys.Error 处理
```go
// 命令执行失败
if err != nil {
    logger.Errorf("命令执行失败: %v, 输出: %v", err, string(output))
    return err  // 向上层传递错误
}
```

**策略**:
- 记录详细错误日志
- 向调用者返回原始错误
- 不吞噬错误

### flagutil.Error 处理
```go
default:
    panic(fmt.Sprintf("不支持的类型: %s", field.Type))
```

**策略**:
- 遇到不支持的类型 panic(启动时崩溃)
- 强制开发者在启动时发现问题
- 避免运行时未知类型错误

---

## 📖 依赖关系

### 外部依赖

```
util/
├── flag/
│   ├── flag (标准库) - 命令行参数解析
│   ├── fmt (标准库) - 格式化输出
│   └── reflect (标准库) - 反射操作
│
├── response/
│   └── MediaCacheService/models/resp - 响应数据结构
│
└── sys/
    ├── os/exec (标准库) - 执行外部命令
    ├── context (标准库) - 上下文管理
    ├── time (标准库) - 时间类型
    └── MediaCacheService/common/logger - 日志模块
```

### 被依赖关系

```
util 被以下模块依赖:
├── main.go (应用启动)
├── controllers/* (HTTP 控制器)
├── service/* (业务逻辑)
├── storage/* (存储管理)
└── tasks/* (后台任务)
```

---

## 🎯 总结

util 包为 MediaCacheService 提供了三个核心基础设施工具:

1. **flagutil** - 命令行参数解析
   - 使用反射自动注册结构体字段
   - 支持嵌套结构体
   - 类型安全,扩展性好

2. **response** - HTTP 响应标准化
   - 统一 API 响应格式
   - 简化控制器代码
   - 易于维护和扩展

3. **sys** - 系统操作接口
   - 目录大小计算
   - 文件清理功能
   - 接口设计,易于测试

这三个模块设计简洁,职责单一,为整个应用提供了可靠的基础设施支持。