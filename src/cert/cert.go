package cert

import (
// "CSPGSOMF/CertSDK/api/base"
// "CSPGSOMF/CertSDK/api/certapi"
// "mediaCacheService/common/https"
// "mediaCacheService/common/logger"
// "mediaCacheService/remote"
)

// 全局变量
//var (
//	exCertMgr      base.CSPExCertManager
//	externalInfos  []base.CspExSceneInfo
//	externalCaInfos []base.CspExSceneInfo
//	serverInfos    []base.CspExSceneInfo
//	newCertInfo    = https.CertInfo{}
//)
//
//// InitCert 初始化证书SDK，获取证书管理器实例
//func InitCert() {
//	logger.Infof("start subscribe sbg certificate scene")
//
//	err := certapi.CspCertSDKInit()
//	if err != nil {
//		logger.Fatalf("init ex cert sdk failed: %v", err)
//	}
//
//	exCertMgr = certapi.GetExCertManagerInstance()
//}
//
//// InitCertScene 初始化证书场景，清理旧订阅
//func InitCertScene() error {
//	// 定义4个证书场景
//	sceneCaInfo := base.CspExSceneInfo{
//		SceneName:   "sbg_external_ca_certificate",
//		SceneDescCN: "云浏览器外部CA证书",
//		SceneDescEN: "SBG external CA certificate",
//		SceneType:   1,
//		Feature:     0,
//	}
//	sceneDeviceInfo := base.CspExSceneInfo{
//		SceneName:   "sbg_external_device_certificate",
//		SceneDescCN: "云浏览器外部设备证书",
//		SceneDescEN: "SBG external Device Certificate",
//		SceneType:   2,
//		Feature:     0,
//	}
//	serverCaInfo := base.CspExSceneInfo{
//		SceneName:   "sbg_server_ca_certificate",
//		SceneDescCN: "云浏览器服务端CA证书",
//		SceneDescEN: "SBG server CA certificate",
//		SceneType:   1,
//		Feature:     0,
//	}
//	serverDeviceInfo := base.CspExSceneInfo{
//		SceneName:   "sbg_server_device_certificate",
//		SceneDescCN: "云浏览器服务端设备证书",
//		SceneDescEN: "SBG server Device Certificate",
//		SceneType:   2,
//		Feature:     0,
//	}
//
//	// 将场景添加到对应数组
//	externalCaInfos = append(externalCaInfos, sceneCaInfo)
//	externalInfos = append(externalInfos, sceneDeviceInfo)
//	serverInfos = append(serverInfos, serverCaInfo)
//	serverInfos = append(serverInfos, serverDeviceInfo)
//
//	// 清理旧订阅，避免换包重启时残留订阅冲突
//	if err := exCertMgr.UnsubscribeExCert("mediacache-muen",
//		[]base.CspExSceneInfo{sceneCaInfo, serverCaInfo, serverDeviceInfo}); err != nil {
//		return err
//	}
//	if err := exCertMgr.UnsubscribeExCert("mediacache-muenCa",
//		[]base.CspExSceneInfo{sceneDeviceInfo, serverCaInfo, serverDeviceInfo}); err != nil {
//		return err
//	}
//	if err := exCertMgr.UnsubscribeExCert("mediacache",
//		[]base.CspExSceneInfo{sceneCaInfo, sceneDeviceInfo}); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// SubscribeCert 订阅三组证书：外部设备证书、外部CA证书、服务端证书
//func SubscribeCert(server *https.BeegoHttpsServer) error {
//	if err := InitCertScene(); err != nil {
//		return err
//	}
//
//	// 订阅外部设备证书，用于Muen客户端
//	if err := exCertMgr.SubscribeExCert("mediacache-muen", externalInfos,
//		exCertInfoHandler, "/opt/csp/mediacache/"); err != nil {
//		return err
//	}
//
//	// 订阅外部CA证书，用于Muen客户端CA
//	if err := exCertMgr.SubscribeExCert("mediacache-muenCa", externalCaInfos,
//		exCertInfoHandler, "/opt/csp/mediacache/"); err != nil {
//		return err
//	}
//
//	// 订阅服务端证书，用于HTTPS服务器
//	if err := exCertMgr.SubscribeExCert("mediacache", serverInfos,
//		func(certInfo []*base.CspExCertInfo, notifyType int) error {
//			return serverCertInfoHandler(server, certInfo, notifyType)
//		}, "/opt/csp/mediacache/"); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// exCertInfoHandler 外部证书更新回调，更新Muen HTTPS客户端证书
//func exCertInfoHandler(certInfo []*base.CspExCertInfo, notifyType int) error {
//	logger.Infof("get sbg external cert update, try update client")
//
//	for _, info := range certInfo {
//		res, err := exCertMgr.GetExCertPathInfo(info.SceneName)
//		if err != nil {
//			logger.Errorf("get cert path failed: %v", err)
//			continue
//		}
//
//		pwd, err := exCertMgr.GetExCertPrivateKeyPwd(info.SceneName)
//		if err != nil {
//			logger.Errorf("get cert pwd failed: %v", err)
//			continue
//		}
//
//		switch info.SceneName {
//		case "sbg_external_ca_certificate":
//			newCertInfo.CaFile = res.ExCaFile
//		case "sbg_external_device_certificate":
//			newCertInfo.CertFile = res.ExDeviceFilePath
//			newCertInfo.KeyFile = res.ExPrivateKeyFilePath
//			newCertInfo.KeyPwd = pwd
//		}
//	}
//
//	remote.UpdateMuenClientInstance(newCertInfo)
//	return nil
//}
//
//// serverCertInfoHandler 服务端证书更新回调，热更新HTTPS服务器证书
//func serverCertInfoHandler(server *https.BeegoHttpsServer, certInfo []*base.CspExCertInfo, notifyType int) error {
//	logger.Infof("get sbg server cert update, try update server")
//
//	cert := https.CertInfo{}
//
//	for _, info := range certInfo {
//		res, err := exCertMgr.GetExCertPathInfo(info.SceneName)
//		if err != nil {
//			logger.Errorf("get cert path failed: %v", err)
//			continue
//		}
//
//		pwd, err := exCertMgr.GetExCertPrivateKeyPwd(info.SceneName)
//		if err != nil {
//			logger.Errorf("get cert pwd failed: %v", err)
//			continue
//		}
//
//		switch info.SceneName {
//		case "sbg_server_ca_certificate":
//			cert.CaFile = res.ExCaFile
//		case "sbg_server_device_certificate":
//			cert.CertFile = res.ExDeviceFilePath
//			cert.KeyFile = res.ExPrivateKeyFilePath
//			cert.KeyPwd = pwd
//		}
//	}
//
//	server.UpdateCert(cert)
//	return nil
//}
