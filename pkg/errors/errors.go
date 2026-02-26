package errors

import "fmt"

type ErrorCode string

const (
	CodeSuccess           ErrorCode = "SUCCESS"
	CodeBadRequest        ErrorCode = "BAD_REQUEST"
	CodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	CodeForbidden         ErrorCode = "FORBIDDEN"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeInternalError     ErrorCode = "INTERNAL_ERROR"
	CodeValidationError   ErrorCode = "VALIDATION_ERROR"
	CodeDatabaseError     ErrorCode = "DATABASE_ERROR"
	CodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
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

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewWithError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

func NotFound(message string) *AppError {
	return New(CodeNotFound, message)
}

func InternalError(message string) *AppError {
	return New(CodeInternalError, message)
}

func ValidationError(message string) *AppError {
	return New(CodeValidationError, message)
}

func DatabaseError(message string, err error) *AppError {
	return NewWithError(CodeDatabaseError, message, err)
}

func ExternalServiceError(message string, err error) *AppError {
	return NewWithError(CodeExternalServiceError, message, err)
}
