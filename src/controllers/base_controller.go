package controllers

import (
	"encoding/json"
	"mediaCacheService/common/logger"
	"mediaCacheService/models/resp"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/beego/beego/v2/server/web"
)

// BaseController 基础控制器，提供通用 HTTP 处理方法
type BaseController struct {
	web.Controller
}

// RouteInfo 基础控制器返回空路由信息，子类可覆盖
func (c *BaseController) RouteInfo() RouteInfo {
	return RouteInfo{}
}

// QueryParameter 获取 URL 查询参数
func (c *BaseController) QueryParameter(name string) string {
	return c.GetString(name)
}

// PathParameter 获取路径参数
func (c *BaseController) PathParameter(name string) string {
	return c.Ctx.Input.Param(":" + name)
}

// WriteHeaderAndJSON 写入 HTTP 状态码并序列化 JSON 响应体
func (c *BaseController) WriteHeaderAndJSON(status int, v interface{}, contentType string) error {
	if v == nil {
		c.Ctx.ResponseWriter.WriteHeader(status)
		return nil
	}
	c.AddHeader("Content-Type", contentType)
	c.Ctx.ResponseWriter.WriteHeader(status)
	return json.NewEncoder(c.Ctx.ResponseWriter).Encode(v)
}

// AddHeader 添加响应头
func (c *BaseController) AddHeader(header, value string) {
	c.Ctx.Output.Header(header, value)
}

// ResponseWriter 获取响应写入器
func (c *BaseController) ResponseWriter() http.ResponseWriter {
	return c.Ctx.ResponseWriter
}

// Request 获取请求对象
func (c *BaseController) Request() *http.Request {
	return c.Ctx.Request
}

// DownloadFile 提供 HTTP 文件下载，自动处理文件名编码
func (c *BaseController) DownloadFile(file string) {
	if _, err := filepath.Abs(file); err != nil {
		http.ServeFile(c.ResponseWriter(), c.Request(), file)
		return
	}

	fName := filepath.Base(file)
	fn := url.PathEscape(fName)
	if fName == fn {
		fn = "filename=" + fn
	} else {
		fn = "filename=" + fName + "; filename*=utf-8''" + fn
	}
	c.AddHeader("Content-Disposition", "attachment; "+fn)
	c.AddHeader("Content-Description", "File Transfer")
	c.AddHeader("Content-Type", "application/octet-stream")
	c.AddHeader("Content-Transfer-Encoding", "binary")
	c.AddHeader("Expires", "0")
	c.AddHeader("Cache-Control", "must-revalidate")
	c.AddHeader("Pragma", "public")
	http.ServeFile(c.ResponseWriter(), c.Request(), file)
}

// OK 返回 HTTP 200 成功响应
func (c *BaseController) OK(data interface{}) {
	if data == nil {
		data = resp.BaseResponse{
			Code:    0,
			Message: "success",
		}
	}
	err := c.WriteHeaderAndJSON(http.StatusOK, data, "application/json")
	if err != nil {
		logger.Errorf("return output %v failed, turn to InternalServiceError", data)
		c.InternalServiceError()
	}
	respData, _ := json.Marshal(data)
	logger.Infof("response is %s", string(respData))
}

// Failed 返回 HTTP 400 失败响应
func (c *BaseController) Failed(data resp.BaseResponse) {
	err := c.WriteHeaderAndJSON(http.StatusBadRequest, data, "application/json")
	if err != nil {
		logger.Errorf("return output %v failed, turn to InternalServiceError", data)
		c.InternalServiceError()
	}
}

// InternalServiceError 返回 HTTP 500 内部错误响应
func (c *BaseController) InternalServiceError() {
	r := resp.BaseResponse{Code: -1, Message: "IMessage:服务内部错误"}
	_ = c.WriteHeaderAndJSON(http.StatusInternalServerError, r, "application/json")
}
