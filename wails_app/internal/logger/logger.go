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

// LogType 定义日志类型
type LogType string

const (
	// LogTypeOperation 用于用户操作
	LogTypeOperation LogType = "operation"
	// LogTypeRunning 用于系统运行状态
	LogTypeRunning LogType = "running"
)

// Logger 处理应用程序日志记录
type Logger struct {
	ctx           context.Context
	operationFile *os.File
	runningFile   *os.File
	mutex         sync.Mutex
	logDir        string
}

var globalLogger *Logger
var once sync.Once

// Init 初始化全局日志记录器
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

		// 打开日志文件
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

// GetInstance 返回全局日志记录器实例
func GetInstance() *Logger {
	return globalLogger
}

// Log 记录一条消息
func (l *Logger) Log(logType LogType, level string, format string, args ...interface{}) {
	if l == nil {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)

	// 写入文件
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

	// 发送到前端
	if l.ctx != nil {
		runtime.EventsEmit(l.ctx, "log", map[string]string{
			"type":    string(logType),
			"level":   level,
			"message": msg,
			"time":    timestamp,
		})
	}

	// 同时打印到标准输出以进行调试
	fmt.Print(logEntry)
}

// Info 记录一条信息消息
func Info(logType LogType, format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Log(logType, "INFO", format, args...)
	}
}

// Error 记录一条错误消息
func Error(logType LogType, format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Log(logType, "ERROR", format, args...)
	}
}

// Warn 记录一条警告消息
func Warn(logType LogType, format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Log(logType, "WARN", format, args...)
	}
}
