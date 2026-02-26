package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(CodeBadRequest, "test message")
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "test message", err.Message)
	assert.Nil(t, err.Err)
}

func TestNewWithError(t *testing.T) {
	originalErr := errors.New("original error")
	err := NewWithError(CodeInternalError, "test message", originalErr)
	assert.Equal(t, CodeInternalError, err.Code)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, originalErr, err.Err)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "without wrapped error",
			err:      New(CodeNotFound, "not found"),
			expected: "not found",
		},
		{
			name:     "with wrapped error",
			err:      NewWithError(CodeInternalError, "internal error", errors.New("database down")),
			expected: "internal error: database down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := NewWithError(CodeDatabaseError, "db error", originalErr)
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name         string
		fn           func() *AppError
		expectedCode ErrorCode
	}{
		{
			name:         "BadRequest",
			fn:           func() *AppError { return BadRequest("bad request") },
			expectedCode: CodeBadRequest,
		},
		{
			name:         "NotFound",
			fn:           func() *AppError { return NotFound("not found") },
			expectedCode: CodeNotFound,
		},
		{
			name:         "InternalError",
			fn:           func() *AppError { return InternalError("internal error") },
			expectedCode: CodeInternalError,
		},
		{
			name:         "ValidationError",
			fn:           func() *AppError { return ValidationError("validation error") },
			expectedCode: CodeValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			assert.Equal(t, tt.expectedCode, err.Code)
		})
	}
}

func TestDatabaseError(t *testing.T) {
	originalErr := errors.New("db connection failed")
	err := DatabaseError("db error", originalErr)
	assert.Equal(t, CodeDatabaseError, err.Code)
	assert.Equal(t, "db error", err.Message)
	assert.Equal(t, originalErr, err.Err)
}

func TestExternalServiceError(t *testing.T) {
	originalErr := errors.New("api timeout")
	err := ExternalServiceError("service error", originalErr)
	assert.Equal(t, CodeExternalServiceError, err.Code)
	assert.Equal(t, "service error", err.Message)
	assert.Equal(t, originalErr, err.Err)
}
