package remote

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"mediaCacheService/common/https"
	"mediaCacheService/common/logger"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	muenClientMu sync.RWMutex
	muenClient   *http.Client
)

// InitMuenClient 初始化基础Muen HTTP客户端（无证书）
func InitMuenClient() {
	muenClientMu.Lock()
	defer muenClientMu.Unlock()
	muenClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	logger.Info("muen client initialized (no cert)")
}

// UpdateMuenClientInstance 使用新证书更新Muen HTTPS客户端
func UpdateMuenClientInstance(certInfo https.CertInfo) {
	client, err := NewHttpsMuenClient(certInfo)
	if err != nil {
		logger.Error("update muen client failed: %v", err)
		return
	}
	muenClientMu.Lock()
	muenClient = client
	muenClientMu.Unlock()
	logger.Info("muen client updated with new cert")
}

// NewHttpsMuenClient 根据证书信息创建Muen HTTPS客户端
func NewHttpsMuenClient(certInfo https.CertInfo) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certInfo.CertFile, certInfo.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load muen client key pair failed: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if certInfo.CaFile != "" {
		caCert, err := os.ReadFile(certInfo.CaFile)
		if err != nil {
			return nil, fmt.Errorf("read muen ca file failed: %w", err)
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCert)
		tlsCfg.RootCAs = pool
	}

	transport := &http.Transport{TLSClientConfig: tlsCfg}
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

// GetMuenClientInstance 获取当前Muen客户端实例
func GetMuenClientInstance() *http.Client {
	muenClientMu.RLock()
	defer muenClientMu.RUnlock()
	return muenClient
}
