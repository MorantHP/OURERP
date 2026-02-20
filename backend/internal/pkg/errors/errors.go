package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用错误
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Details    any    `json:"details,omitempty"`
	Stack      string `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 预定义错误码
const (
	CodeSuccess          = "SUCCESS"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeValidationError  = "VALIDATION_ERROR"
	CodeDatabaseError    = "DATABASE_ERROR"
	CodeCacheError       = "CACHE_ERROR"
	CodeExternalAPIError = "EXTERNAL_API_ERROR"
)

// NewAppError 创建应用错误
func NewAppError(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails 添加详情
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// 预定义错误

// BadRequest 错误请求
func BadRequest(message string) *AppError {
	return NewAppError(CodeBadRequest, message, http.StatusBadRequest)
}

// Unauthorized 未授权
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "未授权访问"
	}
	return NewAppError(CodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden 禁止访问
func Forbidden(message string) *AppError {
	if message == "" {
		message = "禁止访问"
	}
	return NewAppError(CodeForbidden, message, http.StatusForbidden)
}

// NotFound 未找到
func NotFound(resource string) *AppError {
	return NewAppError(CodeNotFound, fmt.Sprintf("%s不存在", resource), http.StatusNotFound)
}

// Conflict 冲突
func Conflict(message string) *AppError {
	return NewAppError(CodeConflict, message, http.StatusConflict)
}

// InternalError 内部错误
func InternalError(message string) *AppError {
	return NewAppError(CodeInternalError, message, http.StatusInternalServerError)
}

// ValidationError 验证错误
func ValidationError(message string, details any) *AppError {
	return NewAppError(CodeValidationError, message, http.StatusBadRequest).WithDetails(details)
}

// DatabaseError 数据库错误
func DatabaseError(message string) *AppError {
	return NewAppError(CodeDatabaseError, message, http.StatusInternalServerError)
}

// ServiceUnavailable 服务不可用
func ServiceUnavailable(message string) *AppError {
	return NewAppError(CodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

// ExternalAPIError 外部API错误
func ExternalAPIError(service string, err error) *AppError {
	return NewAppError(CodeExternalAPIError,
		fmt.Sprintf("%s服务调用失败: %v", service, err),
		http.StatusBadGateway)
}

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 获取应用错误
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return InternalError(err.Error())
}

// 常用错误
var (
	ErrUserNotFound       = NotFound("用户")
	ErrTenantNotFound     = NotFound("租户")
	ErrShopNotFound       = NotFound("店铺")
	ErrProductNotFound    = NotFound("商品")
	ErrOrderNotFound      = NotFound("订单")
	ErrWarehouseNotFound  = NotFound("仓库")
	ErrInventoryNotFound  = NotFound("库存")
	ErrSupplierNotFound   = NotFound("供应商")
	ErrAlertRuleNotFound  = NotFound("预警规则")

	ErrInvalidCredentials = Unauthorized("邮箱或密码错误")
	ErrTokenExpired       = Unauthorized("登录已过期，请重新登录")
	ErrTokenInvalid       = Unauthorized("无效的认证信息")
	ErrPermissionDenied   = Forbidden("权限不足")

	ErrEmailExists        = Conflict("邮箱已被注册")
	ErrTenantCodeExists   = Conflict("租户编码已存在")

	ErrNoTenantSelected   = BadRequest("请选择账套")
	ErrTenantDisabled     = Forbidden("租户已禁用")

	ErrUserNotApproved    = Forbidden("账户待审核，请等待管理员审核")
	ErrUserDisabled       = Forbidden("账户已被禁用")
)

// Wrap 包装错误
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:       appErr.Code,
			Message:    fmt.Sprintf("%s: %s", message, appErr.Message),
			StatusCode: appErr.StatusCode,
			Details:    appErr.Details,
		}
	}
	return InternalError(fmt.Sprintf("%s: %v", message, err))
}
