package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用错误
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误
var (
	// 通用错误
	ErrBadRequest   = &AppError{Code: "BAD_REQUEST", Message: "请求参数错误", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "未授权访问", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden    = &AppError{Code: "FORBIDDEN", Message: "禁止访问", HTTPStatus: http.StatusForbidden}
	ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "资源不存在", HTTPStatus: http.StatusNotFound}
	ErrInternal     = &AppError{Code: "INTERNAL_ERROR", Message: "服务器内部错误", HTTPStatus: http.StatusInternalServerError}

	// 业务错误
	ErrTenantNotSelected = &AppError{Code: "TENANT_NOT_SELECTED", Message: "请选择账套", HTTPStatus: http.StatusBadRequest}
	ErrTenantNotFound    = &AppError{Code: "TENANT_NOT_FOUND", Message: "账套不存在", HTTPStatus: http.StatusNotFound}
	ErrProductNotFound   = &AppError{Code: "PRODUCT_NOT_FOUND", Message: "商品不存在", HTTPStatus: http.StatusNotFound}
	ErrProductDuplicate  = &AppError{Code: "PRODUCT_DUPLICATE", Message: "商品编码已存在", HTTPStatus: http.StatusBadRequest}
	ErrOrderNotFound     = &AppError{Code: "ORDER_NOT_FOUND", Message: "订单不存在", HTTPStatus: http.StatusNotFound}
	ErrInventoryNotEnough = &AppError{Code: "INVENTORY_NOT_ENOUGH", Message: "库存不足", HTTPStatus: http.StatusBadRequest}
)

// NewAppError 创建应用错误
func NewAppError(code, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Wrap 包装错误
func Wrap(err error, code, message string, httpStatus int) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// WrapInternal 包装内部错误
func WrapInternal(err error, message string) *AppError {
	return Wrap(err, "INTERNAL_ERROR", message, http.StatusInternalServerError)
}

// IsAppError 检查是否是应用错误
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
