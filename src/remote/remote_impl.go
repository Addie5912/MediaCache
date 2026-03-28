package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"mediaCacheService/common/conf"
	"mediaCacheService/common/logger"
	"mediaCacheService/storage"
	"net/http"
	"os"
	"strconv"
	"time"
)

// remoteImpl 远程服务实现
type remoteImpl struct {
	httpClient *http.Client
}

// NewRemote 创建远程服务实例
func NewRemote() Remote {
	timeout := time.Duration(conf.GlobalConfig.HTTPTimeout) * time.Second
	return &remoteImpl{
		httpClient: &http.Client{Timeout: timeout},
	}
}

// GetVideo 从MUEN媒体服务器获取视频流
func (r *remoteImpl) GetVideo(videoPath string) (io.ReadCloser, *storage.FileInfo, error) {
	baseURL := os.Getenv("MUEN_MEDIA_URL_PREFIX")
	if baseURL == "" {
		return nil, nil, fmt.Errorf("MUEN_MEDIA_URL_PREFIX not configured")
	}
	url := baseURL + "/" + videoPath
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("get video from MUEN failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("MUEN returned status %d", resp.StatusCode)
	}
	contentLength := resp.Header.Get("Content-Length")
	size := int64(0)
	if contentLength != "" {
		size, _ = strconv.ParseInt(contentLength, 10, 64)
	}
	fileInfo := &storage.FileInfo{
		Path:      videoPath,
		Size:      strconv.FormatInt(size, 10),
		HasCached: false,
	}
	logger.Info("GetVideo from MUEN: %s", videoPath)
	return resp.Body, fileInfo, nil
}

// imeiValidateResponse GIDS鉴权响应结构
type imeiValidateResponse struct {
	Result bool   `json:"result"`
	Msg    string `json:"msg"`
}

// PostValidateIMEI 调用GIDS服务验证IMEI
func (r *remoteImpl) PostValidateIMEI(IMEI string, CheckType string) (bool, error) {
	gidsAddr, err := r.GetGIDSAddress()
	if err != nil {
		return false, fmt.Errorf("get GIDS address failed: %w", err)
	}
	url := fmt.Sprintf("%s/validate?imei=%s&checkType=%s", gidsAddr, IMEI, CheckType)
	resp, err := r.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return false, fmt.Errorf("IMEI validation request failed: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var result imeiValidateResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, fmt.Errorf("decode GIDS response failed: %w", err)
		}
		return result.Result, nil
	case http.StatusUnauthorized:
		return false, fmt.Errorf("IMEI validation not pass")
	case http.StatusInternalServerError:
		return false, fmt.Errorf("GIDS internal server error")
	default:
		return false, fmt.Errorf("GIDS returned unexpected status %d", resp.StatusCode)
	}
}

// GetGIDSAddress 通过CSE服务发现获取GIDS服务地址
func (r *remoteImpl) GetGIDSAddress() (string, error) {
	// TODO: 通过CSE服务发现获取GIDS实例地址
	addr := os.Getenv("GIDS_ADDRESS")
	if addr == "" {
		return "", fmt.Errorf("GIDS_ADDRESS not configured")
	}
	return addr, nil
}
