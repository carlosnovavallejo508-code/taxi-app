package response

import (
    "github.com/gin-gonic/gin"
)

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Code    string      `json:"code,omitempty"`
    Message string      `json:"message,omitempty"`
}

func Success(c *gin.Context, statusCode int, data interface{}) {
    c.JSON(statusCode, Response{
        Success: true,
        Data:    data,
    })
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(201, Response{
        Success: true,
        Data:    data,
        Message: "Created successfully",
    })
}

func Error(c *gin.Context, statusCode int, message, code string) {
    c.JSON(statusCode, Response{
        Success: false,
        Error:   message,
        Code:    code,
    })
}

func SuccessWithMessage(c *gin.Context, statusCode int, data interface{}, message string) {
    c.JSON(statusCode, Response{
        Success: true,
        Data:    data,
        Message: message,
    })
}
