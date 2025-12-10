package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Level represents logging level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Entry represents a single log entry
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Caller    string            `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides structured logging functionality
type Logger struct {
	level     Level
	output    io.Writer
	entries   chan Entry
	done      chan bool
}

// New creates a new Logger
func New(level Level, output io.Writer) *Logger {
	logger := &Logger{
		level:  level,
		output: output,
		entries: make(chan Entry, 1000),
		done:    make(chan bool),
	}

	// Start background writer
	go logger.writer()

	return logger
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	// Ensure logs directory exists
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: Failed to create logs directory: %v", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	logFile := filepath.Join(logDir, fmt.Sprintf("photo-backup-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Warning: Failed to open log file: %v", err)
		// Fallback to stderr
		return New(INFO, os.Stderr)
	}

	// Create multi-writer to write to both file and stderr
	output := io.MultiWriter(file, os.Stderr)

	return New(INFO, output)
}

// writer writes log entries to the output
func (l *Logger) writer() {
	for entry := range l.entries {
		data, err := json.Marshal(entry)
		if err != nil {
			// Fallback to simple log
			log.Printf("Failed to marshal log entry: %v", err)
			continue
		}

		if _, err := l.output.Write(append(data, '\n')); err != nil {
			log.Printf("Failed to write log entry: %v", err)
		}
	}
	l.done <- true
}

// Close closes the logger
func (l *Logger) Close() {
	close(l.entries)
	<-l.done
}

// shouldLog checks if a log level should be logged
func (l *Logger) shouldLog(level Level) bool {
	return level >= l.level
}

// getCaller returns the caller information
func (l *Logger) getCaller() string {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	parts := strings.Split(fn.Name(), "/")
	funcName := parts[len(parts)-1]

	return fmt.Sprintf("%s:%d (%s)", filepath.Base(file), line, funcName)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Field) {
	if !l.shouldLog(DEBUG) {
		return
	}

	l.log(DEBUG, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Field) {
	if !l.shouldLog(INFO) {
		return
	}

	l.log(INFO, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Field) {
	if !l.shouldLog(WARN) {
		return
	}

	l.log(WARN, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Field) {
	if !l.shouldLog(ERROR) {
		return
	}

	l.log(ERROR, msg, fields...)
}

// log creates and sends a log entry
func (l *Logger) log(level Level, msg string, fields ...Field) {
	entry := Entry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   msg,
		Caller:    l.getCaller(),
	}

	if len(fields) > 0 {
		entry.Fields = make(map[string]interface{})
		for _, field := range fields {
			entry.Fields[field.Key] = field.Value
		}
	}

	l.entries <- entry
}

// HTTPRequest logs an HTTP request
func (l *Logger) HTTPRequest(method, path, clientIP string, statusCode int, duration time.Duration) {
	l.Info("HTTP Request",
		String("method", method),
		String("path", path),
		String("client_ip", clientIP),
		Int("status_code", statusCode),
		Float64("duration_ms", float64(duration.Nanoseconds())/1000000.0),
	)
}

// Auth logs authentication events
func (l *Logger) Auth(username, action string, success bool) {
	level := INFO
	if !success {
		level = WARN
	}

	l.log(level, "Authentication",
		String("username", username),
		String("action", action),
		Bool("success", success),
	)
}

// PhotoOperation logs photo operations
func (l *Logger) PhotoOperation(operation, localID, filename string, userID uint, success bool) {
	l.Info("Photo Operation",
		String("operation", operation),
		String("local_id", localID),
		String("filename", filename),
		Uint("user_id", userID),
		Bool("success", success),
	)
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key    string
	Value  interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Uint creates a uint field
func Uint(key string, value uint) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}
