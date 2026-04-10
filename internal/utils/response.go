package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	})
}

func CreatedResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, APIResponse{
		Status:  statusCode,
		Message: message,
		Error:   err,
	})
}

func BadRequestResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Status:  http.StatusBadRequest,
		Message: message,
	})
}

func UnauthorizedResponse(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Status:  http.StatusUnauthorized,
		Message: message,
	})
}

func ForbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Status:  http.StatusForbidden,
		Message: message,
	})
}

func NotFoundResponse(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Status:  http.StatusNotFound,
		Message: message,
	})
}

func InternalServerErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Status:  http.StatusInternalServerError,
		Message: message,
	})
}

func PaginatedSuccessResponse(c *gin.Context, message string, data interface{}, page, limit int, total int64) {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, PaginatedResponse{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
		Meta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
