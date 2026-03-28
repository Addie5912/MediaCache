package routers

import (
	"mediaCacheService/common/constants"
	"mediaCacheService/common/https"
	"mediaCacheService/common/logger"
	"mediaCacheService/controllers"

	"github.com/beego/beego/v2/server/web"
)

// Init 注册所有控制器路由并启动 HTTP/HTTPS 服务器
func Init() {
	// 获取内外网 IP
	internalIP, err := https.GetLocalIP("", constants.FabricEth)
	if err != nil {
		logger.Warnf("get internal ip failed: %v, fallback to 0.0.0.0", err)
		internalIP = "0.0.0.0"
	}
	externalIP, err := https.GetLocalIP("", constants.ScTrunkEth)
	if err != nil {
		logger.Warnf("get external ip failed: %v, fallback to 0.0.0.0", err)
		externalIP = "0.0.0.0"
	}

	// 创建服务器实例
	internalHTTP := https.NewHttpServer(internalIP, constants.DefaultInternalHTTPPort)
	externalHTTP := https.NewHttpServer(externalIP, constants.DefaultExternalHTTPPort)

	servers := []https.BeegoServer{internalHTTP, externalHTTP}

	// 注册所有控制器
	registerControllers(servers, []controllers.IController{
		&controllers.VideoController{},
	})

	// 启动服务器（非阻塞）
	for _, srv := range servers {
		go srv.Run()
	}
}

// registerControllers 将控制器路由和过滤器注册到所有服务器上
func registerControllers(servers []https.BeegoServer, ctrls []controllers.IController) {
	for _, ctrl := range ctrls {
		info := ctrl.RouteInfo()

		for _, srv := range servers {
			// 注册过滤器
			for action, filterFn := range info.Filters {
				pos := web.BeforeRouter
				if action == controllers.After {
					pos = web.AfterExec
				}
				srv.InsertFilter("/*", pos, filterFn)
			}

			// 注册路由映射
			for path, mapping := range info.RouteMapping {
				srv.Router(path, ctrl, mapping)
			}
		}
	}
}
