package constants

const (
	// 服务名称
	ServiceName = "MCS"

	// 网络接口
	FabricEth  = "bond-base"
	ScTrunkEth = "bond-external"

	// 端口默认值
	DefaultInternalHTTPPort  = 9996
	DefaultInternalHTTPSPort = 9997
	DefaultExternalHTTPPort  = 9990
	DefaultExternalHTTPSPort = 9991

	// 文件扩展名白名单
	ExtTS   = ".ts"
	ExtM3U8 = ".m3u8"
	ExtMP4  = ".mp4"

	// 告警ID
	AlarmIDGetVideoFailed = "300020"

	// 告警抑制时间（分钟）
	AlarmSuppressMinutes = 10

	// CSE初始化最大重试次数
	CSEMaxRetry             = 360
	CSERetryIntervalSeconds = 5

	//EnvAppId AppID
	EnvAppId    = "APPID"
	NODENAME    = "NODENAME"
	NAMESPACE   = "NAMESPACE"
	SERVICENAME = "SERVICENAME"

	EnableHttp = "ENABLE_HTTP"
)
