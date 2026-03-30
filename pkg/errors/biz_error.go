package errors

import stderrors "errors"

// BizError 统一业务错误结构，支持在各层携带错误码并透传到 handler。
type BizError struct {
	Code    int
	Message string
	Cause   error
}

// 
func (e *BizError) Error() string {
	if e == nil {
		return Msg(ErrServer)
	}
	if e.Message != "" {
		return e.Message
	}
	return Msg(e.Code)
}

func (e *BizError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func New(code int) error {
	return &BizError{Code: code, Message: Msg(code)}
}

func NewWithMessage(code int, message string) error {
	return &BizError{Code: code, Message: message}
}

func Wrap(code int, cause error) error {
	return &BizError{Code: code, Message: Msg(code), Cause: cause}
}

// CodeOf 提取错误码；非 BizError 默认归类为服务器错误。
func CodeOf(err error) int {
	if err == nil {
		return ErrSuccess
	}
	var bizErr *BizError
	if stderrors.As(err, &bizErr) {
		return bizErr.Code
	}
	return ErrServer
}

// MessageOf 返回对外可展示消息；BizError 优先返回其 Message。
func MessageOf(err error) string {
	if err == nil {
		return Msg(ErrSuccess)
	}
	var bizErr *BizError
	if stderrors.As(err, &bizErr) {
		return bizErr.Error()
	}
	return err.Error()
}
