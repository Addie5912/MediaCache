package util

import (
	"mediaCacheService/models/resp"
)

// Success 创建标准化的 HTTP 成功响应，包含状态码200和指定的数据负载
func Success(data interface{}) resp.DataResponse {
	return resp.DataResponse{
		BaseResponse: resp.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: data,
	}
}
