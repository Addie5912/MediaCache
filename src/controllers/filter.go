package controllers

import (
	"fmt"
	"mediaCacheService/common/logger"
	"net/http"

	beecontext "github.com/beego/beego/v2/server/web/context"
)

const (
	// FilterConfKey 限流配置维度键名
	FilterConfKey = "serviceName.operationName"
	// retryAfter 限流后建议客户端重试的等待秒数
	retryAfter = 10
	// OverloadedResponse 限流时返回给客户端的错误信息
	OverloadedResponse = `{"code":429,"msg":"server is overloaded, please retry later"}`
)

// OverLoadFilter 过载保护过滤器
// 基于请求路径和方法进行限流控制，超过阈值时返回 429 Too Many Requests
func OverLoadFilter(ctx *beecontext.Context) {
	logger.Infof("OverLoadFilter start")

	// 构造限流维度：请求路径 + HTTP 方法
	_ = fmt.Sprintf("%s/%s", ctx.Request.URL.Path, ctx.Input.Method())

	// TODO: 集成 overloadcontroller SDK 实现限流逻辑
	// dimNameValues := map[string]string{
	//     FilterConfKey: ctx.Request.URL.Path + "/" + ctx.Input.Method(),
	// }
	// isGranted, err := overloadcontroller.Process(dimNameValues)
	// if err != nil {
	//     logger.Errorf("overloadcontroller process failed: %v", err)
	// }
	// if !isGranted {
	//     ctx.ResponseWriter.Header().Add("Retry-After", fmt.Sprintf("%d", retryAfter))
	//     ctx.ResponseWriter.WriteHeader(http.StatusTooManyRequests)
	//     if _, err := ctx.ResponseWriter.Write([]byte(OverloadedResponse)); err != nil {
	//         logger.Errorf("OverLoadFilter write response error: %v", err)
	//     }
	//     return
	// }

	// 默认放行（SDK 未接入时不限流）
	_ = http.StatusTooManyRequests // 保持编译器不报 unused import
}
