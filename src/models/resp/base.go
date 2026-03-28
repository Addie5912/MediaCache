package resp

// BaseResponse 基础响应结构（字段名和JSON标签保持不变）
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// DataResponse 带数据的响应结构（字段名和JSON标签保持不变）
type DataResponse struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *DataResponse {
	return &DataResponse{
		BaseResponse: BaseResponse{Code: 0, Message: "success"},
		Data:         data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, msg string) *BaseResponse {
	return &BaseResponse{Code: code, Message: msg}
}
