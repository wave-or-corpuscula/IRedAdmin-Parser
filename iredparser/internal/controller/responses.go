package controller

import (
	"encoding/json"

	apperrors "iredparser/pkg/errors"
)

type Response struct {
	Success bool           `json:"success"`
	Error   *ErrorResponse `json:"error"`
	Data    any            `json:"data"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *CLIController) sendResponse(data any) {
	resp := Response{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(c.outWriter).Encode(resp)
}

func (c *CLIController) SendError(errType string, code int, err error) {
	resp := Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    code,
			Message: err.Error(),
		},
	}

	json.NewEncoder(c.outWriter).Encode(resp)
}

func (c *CLIController) SendIRedError(err *apperrors.IRedError) {
	c.SendError(string(err.Type), int(err.Code), err.Err)
}
