package controllers

import (
	"io"
	"mediaCacheService/common/constants/retcode"
	"mediaCacheService/common/logger"
	"mediaCacheService/models/resp"
	"mediaCacheService/service"
	"mediaCacheService/storage"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// VideoController 视频控制器
// 路由映射（保持不变）:
//
//	GET /video/*    -> GetVideo
//	GET /download/* -> Download
//	GET /test       -> Test
type VideoController struct {
	BaseController
	videoService service.VideoService
	authService  service.AuthService
}

// Prepare 初始化服务实例
func (c *VideoController) Prepare() {
	//r := remote.NewRemote()
	//alarmSvc := service.NewAlarmService()
	//localStorage := storage.NewLocalStorage("LocalStorage")
	c.videoService = service.NewVideoService()
	c.authService = service.NewAuthService()
}

// GetVideo 处理视频获取请求
// GET /video/*
func (c *VideoController) GetVideo() {
	start := time.Now()
	videoPath := c.PathParameter("splat")
	imei := c.QueryParameter("imei")
	checkType := c.QueryParameter("checkType")

	// IMEI鉴权
	valid, err := c.authService.ValidateIMEI(imei, checkType)
	if err != nil {
		logger.Errorf("ValidateIMEI error: %s", err)
		c.Failed(resp.BaseResponse{Code: retcode.InternalFailed, Message: "IMEI validation err"})
		return
	}
	if !valid {
		c.Failed(resp.BaseResponse{Code: retcode.AuthFailed, Message: "IMEI validation not pass"})
		logger.Errorf("ValidateIMEI error: imei %s not valid", imei)
		return
	}

	reader, fileInfo, err := c.videoService.GetVideo(videoPath)
	if err != nil {
		logger.Error("get video failed, err: %v", err)
		resp := resp.BaseResponse{Code: 1, Message: err.Error()}
		_ = c.WriteHeaderAndJSON(http.StatusBadRequest, resp, "application/json")
		return
	}
	defer reader.Close()
	if fileInfo != nil && fileInfo.Finalizer != nil {
		defer fileInfo.Finalizer()
	}

	c.FormatHTTPHeader(fileInfo)
	c.addCorsHeader()

	_, _ = io.Copy(c.ResponseWriter(), reader)
	elapsed := time.Since(start).Milliseconds()
	logger.Info("file[%s] use time: %dms", videoPath, elapsed)
}

// Download 处理文件下载请求
// GET /download/*
func (c *VideoController) Download() {
	path := c.PathParameter("splat")
	reader, size, err := c.videoService.Download(path)
	if err != nil {
		resp := resp.BaseResponse{Code: 1, Message: err.Error()}
		c.Failed(resp)
		return
	}
	defer reader.Close()

	filename := filepath.Base(path)
	c.AddHeader("Content-Disposition", "attachment; filename="+filename)
	c.Ctx.Output.Header("Content-Length", strings.TrimSpace(strings.Join([]string{}, "")))
	_ = size
	_, _ = io.Copy(c.ResponseWriter(), reader)

	resp := resp.DataResponse{
		BaseResponse: resp.BaseResponse{Code: 200, Message: "download success"},
		Data:         true,
	}
	c.Data["json"] = resp
}

// Test 健康检查接口
// GET /test
func (c *VideoController) Test() {
	resp := resp.DataResponse{
		BaseResponse: resp.BaseResponse{Code: 0, Message: "test success"},
		Data:         true,
	}
	_ = c.WriteHeaderAndJSON(http.StatusOK, resp, "application/json")
}

// FormatHTTPHeader 设置视频响应头
func (c *VideoController) FormatHTTPHeader(fileInfo *storage.FileInfo) {
	if fileInfo == nil {
		return
	}
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	contentType := "application/octet-stream"
	switch ext {
	case ".mp4":
		contentType = "video/mp4"
	case ".ts":
		contentType = "video/mp2t"
	case ".m3u8":
		contentType = "application/vnd.apple.mpegurl"
	}
	c.AddHeader("Content-Type", contentType)
	c.AddHeader("Content-Length", fileInfo.Size)
	c.AddHeader("ETag", `"`+fileInfo.Hash+`"`)
	c.AddHeader("content-md5", fileInfo.Hash)
	c.AddHeader("Last-Modified", fileInfo.ModifiedTime.UTC().Format(http.TimeFormat))
	c.AddHeader("Accept-Ranges", "bytes")
	c.AddHeader("Cache-Control", "public,max-age=31536000")
	c.AddHeader("Vary", "Origin")
}

// RouteInfo 返回VideoController的路由映射信息
func (c *VideoController) RouteInfo() RouteInfo {
	return RouteInfo{
		RouteMapping: map[string]string{
			"/video/*":    "GET:GetVideo",
			"/download/*": "GET:Download",
			"/test":       "GET:Test",
		},
	}
}

// addCorsHeader 添加CORS响应头
func (c *VideoController) addCorsHeader() {
	c.AddHeader("Access-Control-Allow-Origin", "*")
	c.AddHeader("Access-Control-Allow-Credentials", "false")
	c.AddHeader("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.AddHeader("Access-Control-Allow-Headers", "Range, Origin, Content-Type")
	c.AddHeader("Access-Control-Expose-Headers", "Content-Range, Last-Modified, Etag, content-md5, Content-Length, Accept-Ranges, Vary")
}
