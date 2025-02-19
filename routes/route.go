package routes

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"
	"xcodeengine/executor"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidRequest = errors.New("invalid request parameters")
	ErrCodeTooLong    = errors.New("code exceeds maximum length")
)

type ExecutionRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
}

type ExecutionResponse struct {
	Output        string `json:"output"`
	Error         string `json:"error,omitempty"`
	StatusMessage string `json:"status_message"`
	ExecutionTime string `json:"execution_time,omitempty"`
}

type ExecutionService struct {
	maxCodeLen int
}

func NewExecutionService() *ExecutionService {
	return &ExecutionService{
		maxCodeLen: 10000,
	}
}

func (s *ExecutionService) HandleExecute(c *gin.Context, workerPool *executor.WorkerPool) {
	var req ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ExecutionResponse{
			Error:         err.Error(),
			StatusMessage: "Invalid Request Format",
		})
		return
	}

	start := time.Now()
	codeBytes, err := base64.StdEncoding.DecodeString(req.Code)
	if err != nil {
		c.JSON(400, "Failed to decode Base64")
		return
	}

	code := string(codeBytes)
	fmt.Println("Time taken to decode base64: ", time.Since(start))

	// Check code length
	if len(req.Code) > s.maxCodeLen {
		c.JSON(400, ExecutionResponse{
			Error:         ErrCodeTooLong.Error(),
			StatusMessage: "Code Too Long",
		})
		return
	}

	// Execute code using worker pool
	output := workerPool.ExecuteJob(req.Language, code)
	logrus.Println("Request: ", req, "Response: ", output)
	if output.Error != nil {
		c.JSON(400, ExecutionResponse{
			Error:         output.Error.Error(),
			StatusMessage: "Runtime Error",
			Output:        output.Output,
		})
		return
	}

	c.JSON(200, ExecutionResponse{
		Output:        output.Output,
		StatusMessage: "Success",
		ExecutionTime: output.ExecutionTime.String(),
	})
}
