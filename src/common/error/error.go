package error

import "fmt"

// Err 自定义错误类型，实现标准 error 接口
type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	// ErrNotExist 数据不存在错误
	ErrNotExist Err = "data not exist"
)

// IsNotExist 判断是否为数据不存在错误
func IsNotExist(err error) bool {
	return err == ErrNotExist
}

// AppError 应用错误，携带错误码和消息
type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Message)
}

// New 创建新的应用错误
func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// Newf 创建带格式化消息的应用错误
func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}
