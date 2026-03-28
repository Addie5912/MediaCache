package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

// OperateType 操作类型枚举
type OperateType int

const (
	GET      OperateType = iota // 读取操作
	ADD                          // 新增操作
	MOD                          // 修改操作
	DELETE                       // 删除操作
	DOWNLOAD                     // 下载操作
	UPLOAD                       // 上传操作
	UPHOLD                       // 维护操作
)

// AuditLogLevel 审计日志级别枚举
type AuditLogLevel int

const (
	MinorLevel     AuditLogLevel = 0 // 次要操作
	ImportantLevel AuditLogLevel = 1 // 重要操作
	LogLevelAuto   AuditLogLevel = 3 // 自动查询
	LogLevelManual AuditLogLevel = 4 // 手动查询
)

// 审计日志上报 URL
const (
	OpsLog = "cse://AuditLog/plat/audit/v1/logs"
	SecLog = "cse://AuditLog/plat/audit/v1/seculogs"
)

// AuditsInfo 审计信息（简化结构）
type AuditsInfo struct {
	Terminal string
	UserName string
	Detail   string
	DetailZh string
}

// AuditsPara 审计日志参数
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

// auditLogBody 审计日志请求体
type auditLogBody struct {
	Operation   string `json:"operation"`
	Level       int    `json:"level"`
	UserName    string `json:"userName"`
	DateTime    int64  `json:"dateTime"`
	AppName     string `json:"appName"`
	AppID       string `json:"appId"`
	Terminal    string `json:"terminal"`
	ServiceName string `json:"serviceName"`
	Result      int    `json:"result"`
	Detail      string `json:"detail"`
	DetailZH    string `json:"detail_zh"`
}

// operationField 操作字段（双重序列化内层）
type operationField struct {
	OPZH string `json:"OP_ZH"`
	OPEN string `json:"OP_EN"`
}

// AuditsLog 提交审计日志到远程服务
func AuditsLog(auditsPara *AuditsPara, requestURL string) {
	opBytes, err := json.Marshal(operationField{
		OPZH: auditsPara.OperationZH,
		OPEN: auditsPara.OperationEN,
	})
	if err != nil {
		Errorf("marshal operation field failed: %v", err)
		return
	}

	body := auditLogBody{
		Operation:   string(opBytes),
		Level:       int(auditsPara.Level),
		UserName:    auditsPara.Username,
		DateTime:    time.Now().Unix(),
		AppName:     "mediacache",
		AppID:       "mediacache",
		Terminal:    auditsPara.Terminal,
		ServiceName: "MCS",
		Result:      auditsPara.Result,
		Detail:      auditsPara.Detail,
		DetailZH:    auditsPara.DetailZH,
	}

	// 第一次序列化
	bs, err := json.Marshal(body)
	if err != nil {
		Errorf("marshal audit log body failed: %v", err)
		return
	}

	// 第二次序列化为字符串（审计服务端要求接收字符串格式）
	bs2, err := json.Marshal(string(bs))
	if err != nil {
		Errorf("marshal audit log string failed: %v", err)
		return
	}

	// TODO: 通过 GSF gsfapi.NewCspRestInvoker().Invoke() 提交日志
	// 当前使用本地日志记录代替远程提交
	Infof("audit log [%s]: %s", requestURL, fmt.Sprintf("%s", bs2))
}

// AuditsSecAndOpsLog 同时记录安全日志和操作日志
func AuditsSecAndOpsLog(secAuditsPara, opsAuditsPara *AuditsPara) {
	AuditsLog(secAuditsPara, SecLog)
	AuditsLog(opsAuditsPara, OpsLog)
}
