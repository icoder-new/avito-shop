package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
)

const (
	BadRequest          = 400
	UnauthorizedError   = 401
	NotFoundError       = 404
	ValidationError     = 422
	InternalServerError = 500
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	err     error  `json:"-"`
}

func (e *AppError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[%d] %s", e.Code, e.Message))
	if e.err != nil {
		b.WriteString(": " + e.err.Error())
	}
	return b.String()
}

func (e *AppError) Unwrap() error {
	return e.err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		err:     err,
	}
}

func IsAppError(err error) bool {
	var appErr *AppError
	return stderrors.As(err, &appErr)
}

func ErrBadRequest(msg string) *AppError {
	return NewAppError(BadRequest, msg, nil)
}

func ErrUnauthorized(msg string) *AppError {
	return NewAppError(UnauthorizedError, msg, nil)
}

func ErrNotFound(msg string) *AppError {
	return NewAppError(NotFoundError, msg, nil)
}

func ErrInternal(err error) *AppError {
	return NewAppError(InternalServerError, "internal server error", err)
}
