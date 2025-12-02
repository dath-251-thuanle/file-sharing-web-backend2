package utils

import (
	"maps"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeInternal        ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	ErrCodeUserNotFound ErrorCode = "User does not exist or invalid id/email"
	ErrCodeLoginInvalid ErrorCode = "Invalid email or password"

	ErrCodeBearerInvalid ErrorCode = "Invalid or missing authentication token"
	ErrCodeDatabaseError ErrorCode = "Error occured with the database"
	ErrCodeFileNotFound  ErrorCode = "File not found"

	ErrCodeUploadBadRequest       ErrorCode = "Bad Upload request"
	ErrCodeUploadPasswordTooShort ErrorCode = "Password too short"
	ErrCodeUploadFileTooBig       ErrorCode = "File size exceeds the system limit"
	ErrCodeFileExpired            ErrorCode = "File has expired"

	ErrCodeDeleteValidationErr ErrorCode = "You do not have permission to delete this file"

	ErrCodeGetForbidden         ErrorCode = "You don't have permission to access this file"
	ErrCodeUploadBearerRequired ErrorCode = "Bearer token is required for authenticated uploads"

	ErrCodeDownloadBearerRequired  ErrorCode = "This file requires authentication. Please provide a Bearer token"
	ErrCodeDownloadPasswordInvalid ErrorCode = "The file password is incorrect"
	ErrCodeFileLocked              ErrorCode = "File not yet available"

	ErrCodeStatForbidden    ErrorCode = "You don't have permission to view statistics for this file"
	ErrCodeFileStatNotFound ErrorCode = "File not found or statistics not available (anonymous upload)"
	ErrCodeHistoryForbidden ErrorCode = "You don't have permission to view download history for this file"

	ErrCodeAdminUnauthorized ErrorCode = "X-Cron-Secret header is required"
	ErrCodeCleanupNotAdmin   ErrorCode = "You don't have permission to perform cleanup"
	ErrCodeCleanUpLimited    ErrorCode = "Cleanup endpoint is rate limited. Please try again later."

	ErrCodeCantAccessResource     ErrorCode = "You don't have permission to access this resource"
	ErrCodeInvalidMaxMinValidDays ErrorCode = "maxValidityDays must be greater than or equal to minValidityHours"
)

type ReturnStatus struct {
	code ErrorCode
	args map[string]any
}

func (bee *ReturnStatus) Error() ErrorCode {
	return bee.code
}

func ErrIfExists(code ErrorCode, e error) *ReturnStatus {
	if e == nil {
		return nil
	}

	return ResponseMsg(code, e.Error())
}

func (bee *ReturnStatus) IsErr() bool {
	if bee != nil {
		return bee.code != ""
	}

	return false
}

func Response(c ErrorCode) *ReturnStatus {
	return &ReturnStatus{
		code: c,
		args: gin.H{},
	}
}

func ResponseMsg(c ErrorCode, message string) *ReturnStatus {
	return &ReturnStatus{
		code: c,
		args: gin.H{"message": message},
	}
}

func ResponseArgs(c ErrorCode, args map[string]any) *ReturnStatus {
	return &ReturnStatus{
		code: c,
		args: args,
	}
}

func (bee *ReturnStatus) Export(c *gin.Context) {
	code := bee.code
	args := bee.args

	switch code {
	// case ErrCodeOk:
	// 	c.JSON(http.StatusOK, args)

	// case ErrFileCreated:
	// 	out := gin.H{
	// 		"success": true,
	// 		"message": "File uploaded successfully",
	// 	}
	// 	maps.Copy(out, args)
	// 	c.JSON(http.StatusCreated, out)
	case ErrCodeLoginInvalid:
		c.JSON(401, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid email or password",
		})

	case ErrCodeUploadBadRequest, ErrCodeBadRequest:
		out := gin.H{
			"error": "Bad request",
		}
		maps.Copy(out, args)
		c.JSON(400, out)

	case ErrCodeUploadBearerRequired:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Bearer token is required for authenticated uploads",
		})

	case ErrCodeUploadFileTooBig:
		c.JSON(413, gin.H{
			"error":   "Payload too large",
			"message": "File size exceeds the system limit",
		})

	case ErrCodeBearerInvalid:
		c.JSON(401, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or missing authentication token",
		})

	case ErrCodeGetForbidden:
		c.JSON(401, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to access this file",
		})

	case ErrCodeFileNotFound:
		c.JSON(404, gin.H{
			"error":   "Not found",
			"message": "File not found",
		})

	case ErrCodeDeleteValidationErr:
		c.JSON(403, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to delete this file",
		})

	case ErrCodeStatForbidden:
		c.JSON(403, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to view statistics for this file",
		})

	case ErrCodeFileStatNotFound:
		c.JSON(404, gin.H{
			"error":   "Not found",
			"message": "File not found or statistics not available (anonymous upload)",
		})

	case ErrCodeHistoryForbidden:
		c.JSON(403, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to view download history for this file",
		})

	case ErrCodeFileExpired:
		out := gin.H{
			"error": "File expired",
		}
		maps.Copy(out, args)
		c.JSON(410, out)

	case ErrCodeDownloadBearerRequired:
		c.JSON(401, gin.H{
			"error":   "Unauthorized",
			"message": "This file requires authentication. Please provide a Bearer token",
		})

	case ErrCodeDownloadPasswordInvalid:
		c.JSON(403, gin.H{
			"error":   "Incorrect password",
			"message": "The file password is incorrect",
		})

	case ErrCodeFileLocked:
		out := gin.H{
			"error": "File not yet available",
		}
		maps.Copy(out, args)
		c.JSON(423, out)

	case ErrCodeAdminUnauthorized:
		c.JSON(401, gin.H{
			"error":   "Unauthorized",
			"message": "X-Cron-Secret header is required",
		})

	case ErrCodeCleanupNotAdmin:
		c.JSON(403, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to perform cleanup",
		})

	case ErrCodeCleanUpLimited:
		c.JSON(429, gin.H{
			"error":   "Too many requests",
			"message": "Cleanup endpoint is rate limited. Please try again later.",
		})

	case ErrCodeCantAccessResource:
		c.JSON(403, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to access this resource",
		})

	case ErrCodeInvalidMaxMinValidDays:
		c.JSON(401, gin.H{
			"error":   "Validation error",
			"message": "maxValidityDays must be greater than or equal to minValidityHours",
		})

	default:
		out := gin.H{
			"error": "Internal Server Error",
		}
		maps.Copy(out, args)
		c.JSON(500, out)
	}
}

type AppError struct {
	Message string
	Code    ErrorCode
	Err     error
}

type APIResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

func (ae *AppError) Error() string {
	return ""
}

func NewError(message string, code ErrorCode) error {
	return &AppError{
		Message: message,
		Code:    code,
	}
}

func WrapError(err error, message string, code ErrorCode) error {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

func ResponseError(ctx *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		status := httpStatusFromCode(appErr.Code)
		response := gin.H{
			"error": CapitalizeFirst(appErr.Message),
			"code":  appErr.Code,
		}

		if appErr.Err != nil {
			response["detail"] = appErr.Err.Error()
		}

		ctx.JSON(status, response)
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
		"code":  ErrCodeInternal,
	})
}

func ResponseSuccess(ctx *gin.Context, status int, message string, data ...any) {
	resp := APIResponse{
		Status:  "success",
		Message: CapitalizeFirst(message),
	}

	if len(data) > 0 && data[0] != nil {
		if m, ok := data[0].(map[string]any); ok {
			if p, exists := m["pagination"]; exists {
				resp.Pagination = p
			}

			if d, exists := m["data"]; exists {
				resp.Data = d
			} else {
				resp.Data = m
			}
		} else {
			resp.Data = data[0]
		}
	}

	ctx.JSON(status, resp)
}

func ResponseStatusCode(ctx *gin.Context, status int) {
	ctx.Status(status)
}

func ResponseValidator(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusBadRequest, data)
}

func httpStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
