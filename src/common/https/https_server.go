package https

import (
	"fmt"
	"os"
	"path"

	"mediaCacheService/common/logger"

	"github.com/beego/beego/v2/server/web"
)

// BeegoHttpsServer HTTPS 服务器实现，支持证书热更新
type BeegoHttpsServer struct {
	server        *web.HttpServer
	ip            string
	port          int
	CertInfo      CertInfo
	restartChan   chan CertInfo
	needStart     bool
	isServerReady bool
}

// NewHttpsServer 创建 HTTPS 服务器实例
func NewHttpsServer(ip string, port int) *BeegoHttpsServer {
	srv := newBeegoHttpsServer(ip, port)
	return &BeegoHttpsServer{
		server:        srv,
		ip:            ip,
		port:          port,
		restartChan:   make(chan CertInfo, 1),
		needStart:     true,
		isServerReady: false,
	}
}

// newBeegoHttpsServer 创建底层 Beego HTTPS 服务器
func newBeegoHttpsServer(ip string, port int) *web.HttpServer {
	cfg := *web.BeeApp.Cfg
	cfg.Listen.EnableHTTPS = true
	cfg.Listen.EnableHTTP = false
	cfg.Listen.HTTPSAddr = ip
	cfg.Listen.HTTPSPort = port

	srv := web.NewHttpSever()
	srv.Cfg = &cfg
	return srv
}

// Run 异步启动 HTTPS 服务器，监听证书更新通道
func (b *BeegoHttpsServer) Run() {
	go func() {
		for certInfo := range b.restartChan {
			b.updateCertInfo(certInfo)
		}
	}()
}

// Router 注册路由
func (b *BeegoHttpsServer) Router(rootpath string, c web.ControllerInterface, mappingMethods ...string) BeegoServer {
	b.server.Router(rootpath, c, mappingMethods...)
	return b
}

// InsertFilter 插入过滤器（中间件）
func (b *BeegoHttpsServer) InsertFilter(pattern string, pos int, filter web.FilterFunc, opts ...web.FilterOpt) BeegoServer {
	b.server.InsertFilter(pattern, pos, filter, opts...)
	return b
}

// UpdateCert 更新服务端证书，触发热更新流程
func (b *BeegoHttpsServer) UpdateCert(certInfo CertInfo) {
	select {
	case b.restartChan <- certInfo:
	default:
		// 通道已满，丢弃旧的，写入新的
		<-b.restartChan
		b.restartChan <- certInfo
	}
}

// updateCertInfo 处理证书更新：首次启动服务器或触发进程重启
func (b *BeegoHttpsServer) updateCertInfo(certInfo CertInfo) {
	if certInfo.CertFile == "" || certInfo.KeyFile == "" {
		logger.Warnf("incomplete cert info, skip update")
		return
	}

	tlsCfg, err := GetTLS(certInfo, ServerType)
	if err != nil {
		logger.Errorf("build tls config failed: %v", err)
		return
	}

	if b.needStart {
		b.server.Cfg.Listen.HTTPSCertFile = certInfo.CertFile
		b.server.Cfg.Listen.HTTPSKeyFile = certInfo.KeyFile
		b.server.Server.TLSConfig = tlsCfg
		b.needStart = false
		go b.server.Run(fmt.Sprintf("%s:%d", b.ip, b.port))
		logger.Infof("https server started on %s:%d", b.ip, b.port)
	} else {
		// 服务已启动，触发进程重启（退出码 3）
		logger.Infof("cert updated, process will restart")
		close(b.restartChan)
		os.Exit(3)
	}
}

// RunWithPresetCert 使用环境变量预设证书启动 HTTPS 服务器
func (b *BeegoHttpsServer) RunWithPresetCert() {
	sslPath := os.Getenv("SSLPATH")
	if sslPath == "" {
		logger.Warnf("SSLPATH not set, https server will wait for cert via UpdateCert")
		b.Run()
		return
	}

	certInfo := CertInfo{
		CaFile:   path.Join(sslPath, "ca.crt"),
		CertFile: path.Join(sslPath, "tls.crt"),
		KeyFile:  path.Join(sslPath, "tls.key.pwd"),
		KeyPwd:   []byte(os.Getenv("INNER_TLS_PRIVATE_KEY_PWD")),
	}

	tlsCfg, err := GetTLS(certInfo, ServerType)
	if err != nil {
		logger.Errorf("load preset cert failed: %v", err)
		b.Run()
		return
	}

	b.server.Cfg.Listen.HTTPSCertFile = certInfo.CertFile
	b.server.Cfg.Listen.HTTPSKeyFile = certInfo.KeyFile
	b.server.Server.TLSConfig = tlsCfg
	b.needStart = false

	go b.server.Run(fmt.Sprintf("%s:%d", b.ip, b.port))
	logger.Infof("https server started with preset cert on %s:%d", b.ip, b.port)

	// 继续监听后续的证书热更新
	b.Run()
}

// close 关闭 HTTPS 服务器
func (b *BeegoHttpsServer) close() {
	if b.server != nil {
		b.server.Server.Close()
	}
}
