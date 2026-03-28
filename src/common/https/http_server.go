package https

import (
	"fmt"
	"net"
	"os"

	"github.com/beego/beego/v2/server/web"
)

// BeegoServer 服务器统一接口，支持链式调用
type BeegoServer interface {
	Run()
	Router(rootpath string, c web.ControllerInterface, mappingMethods ...string) BeegoServer
	InsertFilter(pattern string, pos int, filter web.FilterFunc, opts ...web.FilterOpt) BeegoServer
}

// BeegoHttpServer HTTP 服务器实现
type BeegoHttpServer struct {
	server *web.HttpServer
	ip     string
	port   int
}

// NewHttpServer 创建 HTTP 服务器实例
func NewHttpServer(ip string, port int) *BeegoHttpServer {
	cfg := *web.BeeApp.Cfg
	cfg.Listen.HTTPAddr = ip
	cfg.Listen.HTTPPort = port
	cfg.Listen.EnableHTTPS = false

	srv := web.NewHttpSever()
	srv.Cfg = &cfg

	return &BeegoHttpServer{
		server: srv,
		ip:     ip,
		port:   port,
	}
}

// Run 启动 HTTP 服务器
func (b *BeegoHttpServer) Run() {
	b.server.Run(fmt.Sprintf("%s:%d", b.ip, b.port))
}

// Router 注册路由
func (b *BeegoHttpServer) Router(rootpath string, c web.ControllerInterface, mappingMethods ...string) BeegoServer {
	b.server.Router(rootpath, c, mappingMethods...)
	return b
}

// InsertFilter 插入过滤器（中间件）
func (b *BeegoHttpServer) InsertFilter(pattern string, pos int, filter web.FilterFunc, opts ...web.FilterOpt) BeegoServer {
	b.server.InsertFilter(pattern, pos, filter, opts...)
	return b
}

// GetLocalIP 从环境变量或默认网卡名获取本地 IPv4 地址
func GetLocalIP(ethEnv, defaultEth string) (string, error) {
	netif := os.Getenv(ethEnv)
	if netif == "" {
		netif = defaultEth
	}
	return getEthIP(netif)
}

// getEthIP 获取指定网卡的 IPv4 地址
func getEthIP(ethName string) (string, error) {
	iface, err := net.InterfaceByName(ethName)
	if err != nil {
		return "", fmt.Errorf("interface %s not found: %w", ethName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("get addrs for %s failed: %w", ethName, err)
	}
	for _, addr := range addrs {
		switch ip := addr.(type) {
		case *net.IPNet:
			if ip.IP.To4() != nil && !ip.IP.IsLoopback() {
				return ip.IP.String(), nil
			}
		case *net.IPAddr:
			if ip.IP.To4() != nil && !ip.IP.IsLoopback() {
				return ip.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IPv4 address found on interface %s", ethName)
}
