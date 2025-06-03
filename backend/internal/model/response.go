package model

import (
	"time"
)

// Response 通用响应
type Response struct {
	Code      int       `json:"code"`              // 状态码
	Message   string    `json:"message,omitempty"` // 消息
	Data      any       `json:"data,omitempty"`    // 数据
	Timestamp time.Time `json:"timestamp"`         // 时间戳
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Code      int       `json:"code"`
	Message   string    `json:"message,omitempty"`
	Data      any       `json:"data,omitempty"`
	Page      int       `json:"page"`
	PageSize  int       `json:"pageSize"`
	Total     int       `json:"total"`
	Timestamp time.Time `json:"timestamp"`
}

// NewResponse 创建新的响应
func NewResponse(code int, message string, data any) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data any) *Response {
	return NewResponse(200, "操作成功", data)
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *Response {
	return NewResponse(code, message, nil)
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse(code int, message string, data any, page, pageSize, total int) *PaginatedResponse {
	return &PaginatedResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		Timestamp: time.Now().UTC(),
	}
}

// NowTimestamp 获取当前时间
func NowTimestamp() time.Time {
	return time.Now().UTC()
}
