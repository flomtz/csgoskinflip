package responses

import (
	"encoding/json"
	"net/http"
	"time"

	"csgoskinflip/server/api/utils/errors"
)

type APIResponse struct {
	Status     string        `json:"status"`
	StatusCode int           `json:"status_code"`
	Content    any           `json:"content"`
	Error      *errors.Error `json:"error"` // We make this a pointer because then we can set it to nil, see below
}

func SuccessResponse(code int, content any) *APIResponse {
	return &APIResponse{
		Status:     "success",
		StatusCode: code,
		Content:    content,
		Error:      nil, // here
	}
}

func ErrorResponse(code int, err *errors.Error) *APIResponse {
	return &APIResponse{
		Status:     "error",
		StatusCode: code,
		Content:    nil,
		Error:      err,
	}
}

func HandleRequest[T any](w http.ResponseWriter, r *http.Request, dataFunc func() (*T, *errors.Error)) {
	startTime := time.Now()

	dataObject, dataFuncErr := dataFunc()

	processTime := time.Since(startTime).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	var responseObject *APIResponse

	if dataFuncErr != nil {
		responseObject = ErrorResponse(500, dataFuncErr)
	} else {
		responseObject = SuccessResponse(200, dataObject)
	}

	json.NewEncoder(w).Encode(responseObject)

	totalTime := time.Since(startTime).Milliseconds()

	logEntry := RequestLogEntry{
		LogLevel:      "INFO",
		Method:        r.Method,
		Path:          r.URL.Path,
		StatusCode:    responseObject.StatusCode,
		ProcessTimeMs: processTime,
		TotalTimeMs:   totalTime,
		Error:         dataFuncErr,
	}

	logger.WriteLogEntry(&logEntry)
}
