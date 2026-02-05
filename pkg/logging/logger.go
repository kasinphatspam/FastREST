package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"fastrest/constant"
	"fastrest/metrics"
)

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

type ConsoleLogger struct {
	mu    sync.Mutex
	level LogLevel
}

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func NewLogger() *ConsoleLogger {
	return &ConsoleLogger{
		level: LevelDebug,
	}
}

func (l *ConsoleLogger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *ConsoleLogger) log(level string, levelNum LogLevel, msg string, fields ...interface{}) {
	if levelNum < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().Format("15:04:05")
	levelColor := l.getLevelColor(level)

	fieldStr := ""
	if len(fields) > 0 {
		fieldMap := make(map[string]interface{})
		for i := 0; i < len(fields)-1; i += 2 {
			if key, ok := fields[i].(string); ok {
				fieldMap[key] = fields[i+1]
			}
		}
		if len(fieldMap) > 0 {
			data, _ := json.Marshal(fieldMap)
			fieldStr = " " + string(data)
		}
	}

	fmt.Printf("%s%s%s | %sLOG%s | %s%-7s%s | %s%s%s%s\n",
		constant.ColorGray, now, constant.ColorReset,
		constant.ColorGray, constant.ColorReset,
		levelColor, level, constant.ColorReset,
		msg, constant.ColorGray, fieldStr, constant.ColorReset)
}

func (l *ConsoleLogger) getLevelColor(level string) string {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return constant.ColorGray
	case "INFO":
		return constant.ColorGreen
	case "WARN":
		return constant.ColorYellow
	case "ERROR":
		return constant.ColorRed
	case "FATAL":
		return constant.ColorRed
	default:
		return constant.ColorWhite
	}
}

func (l *ConsoleLogger) Debug(msg string, fields ...interface{}) {
	l.log("DEBUG", LevelDebug, msg, fields...)
}

func (l *ConsoleLogger) Info(msg string, fields ...interface{}) {
	l.log("INFO", LevelInfo, msg, fields...)
}

func (l *ConsoleLogger) Warn(msg string, fields ...interface{}) {
	l.log("WARN", LevelWarn, msg, fields...)
}

func (l *ConsoleLogger) Error(msg string, fields ...interface{}) {
	l.log("ERROR", LevelError, msg, fields...)
}

func (l *ConsoleLogger) Fatal(msg string, fields ...interface{}) {
	l.log("FATAL", LevelFatal, msg, fields...)
	os.Exit(1)
}

type MetricsLogger struct {
	logger  Logger
	metrics *metrics.Metrics
}

func NewMetricsLogger(logger Logger, m *metrics.Metrics) *MetricsLogger {
	return &MetricsLogger{
		logger:  logger,
		metrics: m,
	}
}

func (l *MetricsLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, fields...)
}

func (l *MetricsLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, fields...)
}

func (l *MetricsLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, fields...)
}

func (l *MetricsLogger) Error(msg string, fields ...interface{}) {
	l.metrics.IncLogCount("error")
	l.logger.Error(msg, fields...)
}

func (l *MetricsLogger) Fatal(msg string, fields ...interface{}) {
	l.metrics.IncLogCount("fatal")
	l.logger.Fatal(msg, fields...)
}
