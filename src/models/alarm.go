package models

// AlarmEvent 告警事件模型
type AlarmEvent struct {
	AlarmID      string
	EventMessage string
	Type         string // "generate" or "clear"
}

// AlarmParamInfo 告警参数信息
type AlarmParamInfo struct {
	ParamName  string `json:"paramName"`
	ParamValue string `json:"paramValue"`
}

// AlarmInfo 告警信息结构
type AlarmInfo struct {
	Location   string `json:"location,omitempty"`
	AppendInfo string `json:"appendInfo,omitempty"`
	AlarmId    string `json:"alarmId,omitempty"`
}

// AlarmResponse 告警响应结构
type AlarmResponse struct {
	Retdesc string      `json:"retdesc,omitempty"`
	Data    []AlarmInfo `json:"data,omitempty"`
	RetCode string      `json:"retcode,omitempty"`
}
