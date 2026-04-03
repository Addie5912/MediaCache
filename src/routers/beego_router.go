package routers

import (
	beego "github.com/beego/beego/v2/server/web"

	"mediaCacheService/common/https"
	"mediaCacheService/common/logger"
	"mediaCacheService/controllers"
)

func registerFilters(server https.BeegoServer, routeInfo controllers.RouteInfo, routePathPre string) {
	// 注册全局过滤器匹配所有路由
	for k, v := range routeInfo.Filters {
		var pos = beego.BeforeExec
		if k == controllers.After {
			pos = beego.AfterExec
		}
		server.InsertFilter("/*", pos, v, beego.WithReturnOnOutput(false))
	}
	// 注册带前缀的过滤器（匹配特定前缀的所有子路由）
	if routePathPre != "" {
		for k, v := range routeInfo.Filters {
			var pos = beego.BeforeExec
			if k == controllers.After {
				pos = beego.AfterExec
			}
			server.InsertFilter(routePathPre+"/*", pos, v, beego.WithReturnOnOutput(false))
		}
	}

}

func registerControllerDirectly(server https.BeegoServer, controller controllers.IController) {
	routeInfo := controller.RouteInfo()

	for k, v := range routeInfo.RouteMapping {
		logger.Infof("beego register route(\"%s\")", k)
		server.Router(k, controller, v)
	}

	registerFilters(server, routeInfo, "")
}

func registerController(servers https.BeegoServer, controller controllers.IController) {
	registerControllerDirectly(servers, controller)
}

func RegisterRouters(server https.BeegoServer) {
	server.InsertFilter("*", beego.BeforeRouter, controllers.OverLoadFilter)
	registerController(server, &controllers.VideoController{})
}
