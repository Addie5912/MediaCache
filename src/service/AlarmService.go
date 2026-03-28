package service

import (
	"encoding/json"
	"fmt"
	"mediaCacheService/common/constants"
	"mediaCacheService/common/logger"
	"os"
	"sync"
	"time"
)

// GenerateOrClearType 告警操作类型（对应 AlarmSDK_GO/api/base.GenerateOrClearType）
type GenerateOrClearType int

const (
	GenerateAlarmType GenerateOrClearType = iota // 产生告警
	ClearAlarmType                               // 清除告警
)

const (
	ValuesLen       = 2                   // 位置信息数组长度
	TimePeriodInit  = 3                   // 初始化超时时间（秒）
	TimePeriodClean = 5                   // 清理超时时间（秒）
	RetryTimes      = 360                 // 重试次数
	AlarmId300020   = "300020"            // 告警ID常量
	maxAlarmListLen = 999                 // 告警列表最大长度
	RespOK          = 200                 // HTTP响应码正常
	POST            = "POST"              // HTTP POST方法
	GetActiveAlarms = "GET_ACTIVE_ALARMS" // 获取活动告警命令
	ResultCode      = "0"                 // 结果码（0表示正常）
)

// AlarmService 告警服务接口
type AlarmService interface {
	SendAlarm(alarmID, EventMessage string)
	ClearAlarm(alarmID, EventMessage string)
}

// AlarmEvent 告警事件结构体
type AlarmEvent struct {
	AlarmID      string              // 告警ID
	EventMessage string              // 事件信息
	Type         GenerateOrClearType // 事件类型
}

// AlarmParamInfo 告警参数信息
type AlarmParamInfo struct {
	ParamName  string `json:"paramName"`  // 告警参数名
	ParamValue string `json:"paramValue"` // 告警参数值
}

// AlarmInfo 告警信息
type AlarmInfo struct {
	Location   string `json:"location,omitempty"`
	AppendInfo string `json:"appendInfo,omitempty"`
	AlarmId    string `json:"alarmId,omitempty"`
}

// AlarmResponse 告警响应
type AlarmResponse struct {
	Retdesc string      `json:"retdesc,omitempty"`
	Data    []AlarmInfo `json:"data,omitempty"`
	RetCode string      `json:"retcode,omitempty"`
}

// alarmServiceImpl 告警服务实现结构体
type alarmServiceImpl struct {
	mu     sync.Mutex       // 互斥锁，线程安全
	alarms map[string]int64 // 告警记录表: 告警ID -> 上报时间戳
}

// 包级别变量
var (
	globalAlarmService *alarmServiceImpl
	alarmEventChanel   chan AlarmEvent
)

func init() {
	globalAlarmService = &alarmServiceImpl{
		alarms: make(map[string]int64),
	}
	alarmEventChanel = make(chan AlarmEvent, maxAlarmListLen)
	// 启动告警事件处理协程
	go globalAlarmService.handleEvent()
}

// NewAlarmService 创建告警服务实例（返回全局单例）
func NewAlarmService() AlarmService {
	return globalAlarmService
}

// SendAlarm 发送告警（通过通道异步处理）
func (a *alarmServiceImpl) SendAlarm(alarmID, eventMessage string) {
	a.sendAlarmEvent(AlarmEvent{
		AlarmID:      alarmID,
		EventMessage: eventMessage,
		Type:         GenerateAlarmType,
	})
}

// ClearAlarm 清除告警（通过通道异步处理）
func (a *alarmServiceImpl) ClearAlarm(alarmID, eventMessage string) {
	a.sendAlarmEvent(AlarmEvent{
		AlarmID:      alarmID,
		EventMessage: eventMessage,
		Type:         ClearAlarmType,
	})
}

// sendAlarmEvent 将告警事件发送到通道，5秒超时避免阻塞
func (a *alarmServiceImpl) sendAlarmEvent(event AlarmEvent) {
	select {
	case alarmEventChanel <- event:
	case <-time.After(TimePeriodClean * time.Second):
		logger.Warnf("alarm event channel timeout, alarm[%s] dropped", event.AlarmID)
	}
}

// handleEvent 告警事件处理协程，持续监听通道
func (a *alarmServiceImpl) handleEvent() {
	for event := range alarmEventChanel {
		switch event.Type {
		case GenerateAlarmType:
			a.sendAlarm(event)
		case ClearAlarmType:
			a.clearAlarm(event)
		}
	}
}

// sendAlarm 具体发送告警逻辑（10分钟内重复告警抑制）
func (a *alarmServiceImpl) sendAlarm(event AlarmEvent) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now().Unix()
	if last, ok := a.alarms[event.AlarmID]; ok {
		if now-last < int64(constants.AlarmSuppressMinutes)*60 {
			logger.Infof("alarm[%s] suppressed (within %d min window)", event.AlarmID, constants.AlarmSuppressMinutes)
			return false
		}
	}

	success := a.reportAlarm(event)
	if success {
		a.alarms[event.AlarmID] = now
	}
	return success
}

// clearAlarm 具体清除告警逻辑
func (a *alarmServiceImpl) clearAlarm(event AlarmEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.alarms[event.AlarmID]; !ok {
		logger.Infof("alarm[%s] not found, skip clear", event.AlarmID)
		return
	}

	a.reportAlarm(event)
	delete(a.alarms, event.AlarmID)
}

// reportAlarm 向告警系统上报告警，最多重试2次，每次间隔10秒
func (a *alarmServiceImpl) reportAlarm(event AlarmEvent) bool {
	const maxRetry = 2
	for i := 0; i <= maxRetry; i++ {
		if i > 0 {
			time.Sleep(10 * time.Second)
		}
		// TODO: 通过 CSP AlarmSDK 上报告警
		// alarmObj := alarmapi.NewAlarm(event.AlarmID, ...)
		// err := a.alarmManager.ReportAlarm(alarmObj)
		if event.Type == GenerateAlarmType {
			logger.Infof("[AlarmService] reportAlarm: alarmID[%s] message[%s]", event.AlarmID, event.EventMessage)
		} else {
			logger.Infof("[AlarmService] clearAlarm: alarmID[%s] message[%s]", event.AlarmID, event.EventMessage)
		}
		return true
	}
	return false
}

// CleanAllActiveAlarm 清除所有活动告警（最多重试360次）
func CleanAllActiveAlarm() bool {
	var alarmMap map[string][]AlarmParamInfo
	var err error

	for i := 0; i < RetryTimes; i++ {
		alarmMap, err = GetAllActiveAlarmFromFMService(AlarmId300020)
		if err == nil {
			break
		}
		logger.Warnf("[AlarmService] GetAllActiveAlarmFromFMService failed(retry %d): %v", i+1, err)
		time.Sleep(TimePeriodInit * time.Second)
	}

	if err != nil {
		logger.Errorf("[AlarmService] CleanAllActiveAlarm failed after %d retries", RetryTimes)
		return false
	}

	sourceIP := os.Getenv("POD_IP")
	for alarmID, params := range alarmMap {
		// 根据源IP过滤本节点的告警
		for _, p := range params {
			if p.ParamName == "sourceip" && p.ParamValue == sourceIP {
				clearHistoryAlarm(alarmID)
				break
			}
		}
	}
	return true
}

// clearHistoryAlarm 清除历史告警
func clearHistoryAlarm(alarmID string) {
	// TODO: 通过 CSP AlarmSDK 清除历史告警
	logger.Infof("[AlarmService] clearHistoryAlarm: alarmID[%s]", alarmID)
}

// GetAllActiveAlarmFromFMService 从FMService获取所有活动告警
func GetAllActiveAlarmFromFMService(alarmIds string) (map[string][]AlarmParamInfo, error) {
	requestParams, err := json.Marshal(map[string]string{
		"alarmIds": alarmIds,
		"cmd":      GetActiveAlarms,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request params failed: %w", err)
	}
	return handlerActivityAlarmData("FMService", requestParams)
}

// handlerActivityAlarmData 处理活动告警数据
func handlerActivityAlarmData(micServiceName string, requestParams []byte) (map[string][]AlarmParamInfo, error) {
	// TODO: 通过 CSP GSF 调用微服务接口获取活动告警
	result := make(map[string][]AlarmParamInfo)
	logger.Infof("[AlarmService] handlerActivityAlarmData: service[%s] params[%s]", micServiceName, string(requestParams))
	return result, nil
}

// OSHttpsGetRequestByCSE 通过CSE发送HTTP请求
func OSHttpsGetRequestByCSE(url string, microServiceName string, method string, body []byte) (string, error) {
	// TODO: 通过 CSP GSF CspRestInvoker 发送请求
	logger.Infof("[AlarmService] OSHttpsGetRequestByCSE: url[%s] service[%s] method[%s]", url, microServiceName, method)
	return "", nil
}
