package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm/logger"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	ErrorWithStack(err error, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// 日志字段构造器
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// SimpleLogger 简单日志实现
type SimpleLogger struct {
	level  LogLevel
	fields []Field
	prefix string
}

var defaultLogger *SimpleLogger

func init() {
	defaultLogger = NewLogger(LevelInfo)
}

// NewLogger 创建日志器
func NewLogger(level LogLevel) *SimpleLogger {
	return &SimpleLogger{
		level:  level,
		fields: make([]Field, 0),
	}
}

// SetLevel 设置日志级别
func (l *SimpleLogger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *SimpleLogger) log(level LogLevel, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR"}[level]
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// 构建日志行
	line := fmt.Sprintf("[%s] %s %s", timestamp, levelStr, msg)

	// 添加字段
	allFields := append(l.fields, fields...)
	for _, f := range allFields {
		line += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}

	// 输出
	if level >= LevelError {
		log.Println(line)
	} else {
		log.Println(line)
	}
}

func (l *SimpleLogger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

func (l *SimpleLogger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

func (l *SimpleLogger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

func (l *SimpleLogger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

func (l *SimpleLogger) ErrorWithStack(err error, fields ...Field) {
	l.Error(err.Error(), append(fields, String("stack", "stack trace placeholder"))...)
}

func (l *SimpleLogger) With(fields ...Field) Logger {
	newLogger := &SimpleLogger{
		level:  l.level,
		fields: append(l.fields, fields...),
	}
	return newLogger
}

func (l *SimpleLogger) WithContext(ctx context.Context) Logger {
	// 可以从context提取traceID等信息
	return l
}

// 全局日志函数

func Debug(msg string, fields ...Field) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}

func ErrorWithStack(err error, fields ...Field) {
	defaultLogger.ErrorWithStack(err, fields...)
}

func With(fields ...Field) Logger {
	return defaultLogger.With(fields...)
}

func WithContext(ctx context.Context) Logger {
	return defaultLogger.WithContext(ctx)
}

// SetGlobalLevel 设置全局日志级别
func SetGlobalLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// GormLogger GORM日志适配器
type GormLogger struct {
	logger.Interface
}

func NewGormLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
}

// RequestLogger 请求日志中间件所需的结构
type RequestLog struct {
	StartTime   time.Time
	Method      string
	Path        string
	Query       string
	IP          string
	UserAgent   string
	StatusCode  int
	Latency     time.Duration
	RequestBody string
	Error       string
}

func (l *RequestLog) String() string {
	return fmt.Sprintf(
		"[%s] %s %s %s %d %v",
		l.StartTime.Format("2006-01-02 15:04:05"),
		l.Method,
		l.Path,
		l.IP,
		l.StatusCode,
		l.Latency,
	)
}

// FileWriter 文件日志写入器
type FileWriter struct {
	file *os.File
}

func NewFileWriter(path string) (*FileWriter, error) {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWriter{file: file}, nil
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

func (w *FileWriter) Close() error {
	return w.file.Close()
}
