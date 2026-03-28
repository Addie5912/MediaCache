package controllers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	beecontext "github.com/beego/beego/v2/server/web/context"
)

// newBeeContext 构造用于测试的 Beego 上下文
func newBeeContext(method, path string) (*beecontext.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	req.URL = &url.URL{Path: path}
	rw := httptest.NewRecorder()
	ctx := beecontext.NewContext()
	ctx.Reset(rw, req)
	return ctx, rw
}

// TestOverLoadFilter_Passthrough 验证默认情况下过滤器放行所有请求
func TestOverLoadFilter_Passthrough(t *testing.T) {
	ctx, rw := newBeeContext(http.MethodGet, "/video/test.mp4")

	// 不应 panic，且默认不拦截请求
	OverLoadFilter(ctx)

	// 默认放行，响应码应为 200（未被过滤器写入）
	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}
}

// TestOverLoadFilter_PathAndMethod 验证过滤器可以正确读取路径和方法
func TestOverLoadFilter_PathAndMethod(t *testing.T) {
	paths := []string{"/video/sample.ts", "/download/file.mp4", "/test"}
	methods := []string{http.MethodGet, http.MethodGet, http.MethodGet}

	for i, p := range paths {
		ctx, _ := newBeeContext(methods[i], p)
		// 验证过滤器不 panic
		OverLoadFilter(ctx)
	}
}

// TestFilterAction_Constants 验证 FilterAction 常量值
func TestFilterAction_Constants(t *testing.T) {
	if Before != "before" {
		t.Errorf("expected Before=\"before\", got %q", Before)
	}
	if After != "after" {
		t.Errorf("expected After=\"after\", got %q", After)
	}
}

// TestRouteInfo_EmptyDefault 验证 RouteInfo 默认为空结构体
func TestRouteInfo_EmptyDefault(t *testing.T) {
	ri := RouteInfo{}
	if ri.RouteMapping != nil {
		t.Errorf("expected nil RouteMapping")
	}
	if ri.Filters != nil {
		t.Errorf("expected nil Filters")
	}
}
