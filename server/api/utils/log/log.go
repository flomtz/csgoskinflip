package log

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogObject interface {
	SetTimestamp(timestamp string)
}

type Logger struct {
	logger *log.Logger
}

func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

var (
	initLoggers     []func()
	requestedLogDir string
	SetupDone       bool
)

func RegisterInit(f func()) {
	initLoggers = append(initLoggers, f)
}

func Setup(logDir string) {
	requestedLogDir = logDir

	for _, f := range initLoggers {
		f()
	}

	SetupDone = true
}

func NewLogger(logfileName string) (*Logger, error) {
	err := os.MkdirAll(requestedLogDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	path := filepath.Join(requestedLogDir, logfileName)

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &Logger{logger: log.New(logFile, "", 0)}, nil
}

func (lg *Logger) WriteLogEntry(entry LogObject) {
	if SetupDone {
		entry.SetTimestamp(getTimestamp())
		logData, _ := json.Marshal(entry)

		lg.logger.Println(string(logData))
	} else {
		fmt.Println("Assuming test, not logging.")
	}
}
