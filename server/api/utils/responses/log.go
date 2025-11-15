package responses

import (
	"fmt"

	"csgoskinflip/server/api/utils/errors"
	"csgoskinflip/server/api/utils/log"
)

var logger *log.Logger

func setupRequestLogger() {
	var err error
	logger, err = log.NewLogger("http.log")
	if err != nil {
		fmt.Printf("failed to set up request logger: %v\n", err)
	}
}

func init() {
	log.RegisterInit(setupRequestLogger)
}

type RequestLogEntry struct {
	LogLevel      string        `json:"log_level"`
	Time          string        `json:"time"`
	StatusCode    int           `json:"status_code"`
	RequestId     string        `json:"request_id"`
	TraceId       string        `json:"trace_id"`
	Method        string        `json:"method"`
	Path          string        `json:"path"`
	ProcessTimeMs int64         `json:"process_time_ms"`
	TotalTimeMs   int64         `json:"total_time_ms"`
	Error         *errors.Error `json:"error"`
}

func (wsLogEntry *RequestLogEntry) SetTimestamp(timestamp string) {
	wsLogEntry.Time = timestamp
}
