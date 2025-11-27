package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/net/context"
)

// LogType defines the type of log
type LogType string

const (
	// LogTypeOperation for user operations
	LogTypeOperation LogType = "operation"
	// LogTypeRunning for system running status
	LogTypeRunning LogType = "running"
)

// Logger handles application logging
type Logger struct {
	ctx           context.Context
	operationFile *os.File
	runningFile   *os.File
	mutex         sync.Mutex
	logDir        string
}

var globalLogger *Logger
var once sync.Once

// Init initializes the global logger
func Init(ctx context.Context) error {
	var err error
	once.Do(func() {
		cwd, _ := os.Getwd()
		logDir := filepath.Join(cwd, "logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Error creating log directory: %v\n", err)
			return
		}

		l := &Logger{
			ctx:    ctx,
			logDir: logDir,
		}

		// Open log files
		l.operationFile, err = os.OpenFile(filepath.Join(logDir, "operation.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}

		l.runningFile, err = os.OpenFile(filepath.Join(logDir, "running.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}

		globalLogger = l
	})
	return err
}

// GetInstance returns the global logger instance
func GetInstance() *Logger {
	return globalLogger
}

// Log logs a message
func (l *Logger) Log(logType LogType, level string, format string, args ...interface{}) {
	if l == nil {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)

	// Write to file
	var err error
	if logType == LogTypeOperation {
		if l.operationFile != nil {
			_, err = l.operationFile.WriteString(logEntry)
		}
	} else {
		if l.runningFile != nil {
			_, err = l.runningFile.WriteString(logEntry)
		}
	}

	if err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}

	// Emit to frontend
	if l.ctx != nil {
		runtime.EventsEmit(l.ctx, "log", map[string]string{
			"type":    string(logType),
			"level":   level,
			"message": msg,
			"time":    timestamp,
		})
	}

	// Also print to stdout for debugging
	fmt.Print(logEntry)
}

// Info logs an info message
func Info(logType LogType, format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Log(logType, "INFO", format, args...)
	}
}

// Error logs an error message
func Error(logType LogType, format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Log(logType, "ERROR", format, args...)
	}
}

// Warn logs a warning message
func Warn(logType LogType, format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Log(logType, "WARN", format, args...)
	}
}
