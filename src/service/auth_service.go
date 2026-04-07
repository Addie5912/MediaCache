package service

import (
	"mediaCacheService/common/logger"
	"mediaCacheService/remote"
	"time"
)

// AuthService 鉴权服务接口
type AuthService interface {
	ValidateIMEI(imei string, checkType string) (bool, error)
}

// authServiceImpl 鉴权服务实现结构体
type authServiceImpl struct {
	remote remote.Remote // 远程服务接口
}

// NewAuthService 创建鉴权服务实例（依赖注入）
//func NewAuthService(r remote.Remote) AuthService {
//	return &authServiceImpl{remote: r}
//}

func NewAuthService() *authServiceImpl {
	return &authServiceImpl{
		remote: remote.NewRemote(),
	}
}

// ValidateIMEI 验证设备IMEI是否有效
func (a *authServiceImpl) ValidateIMEI(imei string, checkType string) (bool, error) {
	start := time.Now()
	logger.Infof("[AuthService] start to validate IMEI: %s, checkType: %s", imei, checkType)

	// 尼日利亚空imei情形
	//if imei == "" {
	//	return false, fmt.Errorf("[AuthService] IMEI is empty")
	//}

	result, err := a.remote.PostValidateIMEI(imei, checkType)
	if err != nil {
		logger.Errorf("[AuthService] IMEI validation failed, err: %v", err)
		return false, err
	}

	elapsed := time.Since(start)
	logger.Infof("[AuthService] IMEI validation completed, result: %v, elapsed: %v", result, elapsed)
	return result, nil
}
