package https

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// TLSType 服务类型枚举
type TLSType int

const (
	ServerType TLSType = iota
	ClientType
)

// CertInfo 证书信息
type CertInfo struct {
	CaFile   string
	CertFile string
	KeyFile  string
	KeyPwd   []byte
}

// verify 证书验证器
type verify struct {
	rootCA *x509.CertPool
}

// GetTLS 根据证书信息和服务类型生成 TLS 配置
func GetTLS(info CertInfo, tlsType TLSType) (*tls.Config, error) {
	caPool, cert, err := getCert(info)
	if err != nil {
		return nil, err
	}

	v := &verify{rootCA: caPool}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		NextProtos:         []string{"h2", "http/1.1"},
		InsecureSkipVerify: true, // 使用自定义验证函数
		VerifyConnection:   v.verifyConnection,
	}

	if cert != nil {
		cfg.Certificates = []tls.Certificate{*cert}
	}

	switch tlsType {
	case ServerType:
		if caPool != nil {
			cfg.ClientCAs = caPool
			cfg.ClientAuth = tls.RequireAndVerifyClientCert
		}
	case ClientType:
		if caPool != nil {
			cfg.RootCAs = caPool
		}
	}

	return cfg, nil
}

// getCert 加载证书文件，支持加密私钥
func getCert(info CertInfo) (*x509.CertPool, *tls.Certificate, error) {
	var caPool *x509.CertPool
	var tlsCert *tls.Certificate

	if info.CertFile != "" && info.KeyFile != "" {
		certPEM, err := defaultLoadFile(info.CertFile)
		if err != nil {
			return nil, nil, fmt.Errorf("read cert file failed: %w", err)
		}
		tlsKey, err := defaultLoadFile(info.KeyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("read key file failed: %w", err)
		}

		// 解密加密的私钥
		if len(info.KeyPwd) != 0 {
			privBlock, _ := pem.Decode(tlsKey)
			if privBlock == nil {
				return nil, nil, fmt.Errorf("failed to decode PEM block from key file")
			}
			decryptedBlock, err := x509.DecryptPEMBlock(privBlock, info.KeyPwd)
			if err != nil {
				return nil, nil, fmt.Errorf("decrypt private key failed: %w", err)
			}
			tlsKey = pem.EncodeToMemory(&pem.Block{
				Type:  privBlock.Type,
				Bytes: decryptedBlock,
			})
		}

		cert, err := tls.X509KeyPair(certPEM, tlsKey)
		if err != nil {
			return nil, nil, fmt.Errorf("load x509 key pair failed: %w", err)
		}
		tlsCert = &cert
	}

	if info.CaFile != "" {
		caPEM, err := defaultLoadFile(info.CaFile)
		if err != nil {
			return nil, nil, fmt.Errorf("read ca file failed: %w", err)
		}
		caPool = x509.NewCertPool()
		caPool.AppendCertsFromPEM(caPEM)
	}

	return caPool, tlsCert, nil
}

// defaultLoadFile 安全加载文件内容（转换为绝对路径）
func defaultLoadFile(filePath string) ([]byte, error) {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(abs)
}

// verifyConnection 综合验证 TLS 连接
func (v *verify) verifyConnection(cs tls.ConnectionState) error {
	if len(cs.PeerCertificates) == 0 {
		return fmt.Errorf("no peer certificate")
	}
	cert := cs.PeerCertificates[0]

	if err := v.checkValidity(cert); err != nil {
		return err
	}
	if err := v.checkBasicConstraints(cert); err != nil {
		return err
	}
	if err := v.checkSignatureAlgorithm(cert); err != nil {
		return err
	}
	if err := v.checkKeyUsage(cert); err != nil {
		return err
	}
	if len(cert.DNSNames) > 0 {
		if err := v.verifyHostname(cert, cs.ServerName); err != nil {
			return err
		}
	}
	return nil
}

// checkValidity 验证证书有效期
func (v *verify) checkValidity(cert *x509.Certificate) error {
	now := time.Now()
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}
	return nil
}

// checkBasicConstraints 确保服务器证书不是 CA 证书
func (v *verify) checkBasicConstraints(cert *x509.Certificate) error {
	if cert.IsCA {
		return fmt.Errorf("server certificate cannot be a CA certificate")
	}
	return nil
}

// checkSignatureAlgorithm 拒绝弱签名算法
func (v *verify) checkSignatureAlgorithm(cert *x509.Certificate) error {
	weakAlgorithms := map[x509.SignatureAlgorithm]bool{
		x509.MD2WithRSA:    true,
		x509.MD5WithRSA:    true,
		x509.DSAWithSHA1:   true,
		x509.ECDSAWithSHA1: true,
		x509.SHA1WithRSA:   true,
	}
	if weakAlgorithms[cert.SignatureAlgorithm] {
		return fmt.Errorf("weak signature algorithm: %v", cert.SignatureAlgorithm)
	}
	return nil
}

// checkKeyUsage 验证密钥用法，拒绝不安全的密钥用法
func (v *verify) checkKeyUsage(cert *x509.Certificate) error {
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		return fmt.Errorf("cert should not allow certificate signing")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		return fmt.Errorf("certificate should not allow CRL signature")
	}
	return nil
}

// verifyHostname 验证主机名或 IP 地址，支持通配符证书
func (v *verify) verifyHostname(cert *x509.Certificate, hostname string) error {
	// 检查是否是 IP 地址
	if ip := net.ParseIP(hostname); ip != nil {
		for _, certIP := range cert.IPAddresses {
			if certIP.Equal(ip) {
				return nil
			}
		}
		return fmt.Errorf("certificate is not valid for IP %s", hostname)
	}
	// 检查 DNS 名称（支持通配符）
	for _, dnsName := range cert.DNSNames {
		if matchHostname(dnsName, hostname) {
			return nil
		}
	}
	return fmt.Errorf("certificate is not valid for host %s", hostname)
}

// matchHostname 支持通配符的主机名匹配
func matchHostname(pattern, hostname string) bool {
	if pattern == hostname {
		return true
	}
	// 通配符匹配：*.example.com 匹配 foo.example.com
	if len(pattern) > 2 && pattern[:2] == "*." {
		suffix := pattern[1:] // .example.com
		if len(hostname) > len(suffix) && hostname[len(hostname)-len(suffix):] == suffix {
			// 确保通配符只匹配一层（不含 .）
			prefix := hostname[:len(hostname)-len(suffix)]
			for _, c := range prefix {
				if c == '.' {
					return false
				}
			}
			return true
		}
	}
	return false
}

// NewTLSConfig 创建基础 TLS 配置（仅使用证书文件，无自定义验证）
func NewTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}