package controllers

import (
	"github.com/beego/beego/v2/server/web"
)

// FilterAction 过滤器执行时机枚举
type FilterAction string

const (
	Before FilterAction = "before" // 路由执行前
	After  FilterAction = "after"  // 路由执行后
)

// RouteInfo 路由信息，描述控制器的路由映射和过滤器
type RouteInfo struct {
	RouteMapping map[string]string              // 路径 -> "HTTP方法:方法名"
	Filters      map[FilterAction]web.FilterFunc // 过滤器映射
}

// IController 控制器接口，所有控制器必须实现
type IController interface {
	web.ControllerInterface
	RouteInfo() RouteInfo
}
