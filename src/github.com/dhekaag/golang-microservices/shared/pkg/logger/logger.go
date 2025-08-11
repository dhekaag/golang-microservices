package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Logger struct {
	*slog.Logger
	config Config
}

type Config struct {
	Level       string `json:"level"`
	Format      string `json:"format"`
	ServiceName string `json:"service_name"`
	Environment string `json:"environment"`
}

// Context keys
type ContextKey string

const (
	RequestIDKey     ContextKey = "request_id"
	UserIDKey        ContextKey = "user_id"
	CorrelationIDKey ContextKey = "correlation_id"
)

// Global logger instance
var globalLogger *Logger

// Color codes for different log levels
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
)

// Custom handler that writes directly to stdout with proper formatting
type PrettyHandler struct {
	writer  io.Writer
	level   slog.Level
	service string
}

func NewPrettyHandler(w io.Writer, opts *slog.HandlerOptions, serviceName string) *PrettyHandler {
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		level = opts.Level.Level()
	}

	return &PrettyHandler{
		writer:  w,
		level:   level,
		service: serviceName,
	}
}

func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	timestamp := time.Now().Format("15:04:05")
	level := formatLevel(r.Level)
	service := fmt.Sprintf("%s%s%s", ColorCyan, h.service, ColorReset)

	// Build the log line
	var parts []string
	parts = append(parts, fmt.Sprintf("%s%s%s", ColorGray, timestamp, ColorReset))
	parts = append(parts, level)
	parts = append(parts, service)
	parts = append(parts, r.Message)

	// Add attributes
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "request_id" {
			parts = append(parts, fmt.Sprintf("%s[%s]%s", ColorBlue, a.Value.String(), ColorReset))
		} else if a.Key == "user_id" {
			parts = append(parts, fmt.Sprintf("%suser:%s%s", ColorGreen, a.Value.String(), ColorReset))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%s", a.Key, a.Value.String()))
		}
		return true
	})

	line := strings.Join(parts, " ") + "\n"
	_, err := h.writer.Write([]byte(line))
	return err
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return h
}

// Initialize logger
func Init(config Config) (*Logger, error) {
	level := parseLevel(config.Level)
	serviceName := fmt.Sprintf("%s[%s]", config.ServiceName, config.Environment)

	var handler slog.Handler

	switch strings.ToLower(config.Format) {
	case "json":
		opts := &slog.HandlerOptions{Level: level}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		// Use our custom pretty handler
		opts := &slog.HandlerOptions{Level: level}
		handler = NewPrettyHandler(os.Stdout, opts, serviceName)
	}

	logger := &Logger{
		Logger: slog.New(handler),
		config: config,
	}

	globalLogger = logger
	return logger, nil
}

func formatLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return fmt.Sprintf("%s[DEBUG]%s", ColorGray, ColorReset)
	case slog.LevelInfo:
		return fmt.Sprintf("%s[INFO] %s", ColorGreen, ColorReset)
	case slog.LevelWarn:
		return fmt.Sprintf("%s[WARN] %s", ColorYellow, ColorReset)
	case slog.LevelError:
		return fmt.Sprintf("%s[ERROR]%s", ColorRed, ColorReset)
	default:
		return fmt.Sprintf("[%s]", level.String())
	}
}

// Get global logger
func Get() *Logger {
	if globalLogger == nil {
		// Default logger if not initialized
		globalLogger, _ = Init(Config{
			Level:       "info",
			Format:      "text",
			ServiceName: "unknown",
			Environment: "dev",
		})
	}
	return globalLogger
}

// Simple logging methods with context
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelInfo, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelWarn, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelError, msg, args...)
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, slog.LevelDebug, msg, args...)
}

// Simple logging methods without context (with prettier formatting)
func (l *Logger) InfoMsg(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) WarnMsg(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

func (l *Logger) ErrorMsg(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

func (l *Logger) DebugMsg(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Specialized logging methods with enhanced formatting
func (l *Logger) HTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	level := slog.LevelInfo
	statusColor := ColorGreen

	if statusCode >= 500 {
		level = slog.LevelError
		statusColor = ColorRed
	} else if statusCode >= 400 {
		level = slog.LevelWarn
		statusColor = ColorYellow
	}

	methodColor := ColorCyan
	if method == "POST" || method == "PUT" || method == "DELETE" {
		methodColor = ColorYellow
	}

	msg := fmt.Sprintf("HTTP %s%s%s %s → %s%d%s (%s)",
		methodColor, method, ColorReset,
		path,
		statusColor, statusCode, ColorReset,
		duration.String(),
	)

	l.logWithContext(ctx, level, msg)
}

func (l *Logger) Database(ctx context.Context, operation string, duration time.Duration, err error) {
	if err != nil {
		l.logWithContext(ctx, slog.LevelError,
			fmt.Sprintf("🔴 DB %s%s%s failed", ColorRed, operation, ColorReset),
			"duration", duration.String(),
			"error", err.Error(),
		)
	} else {
		l.logWithContext(ctx, slog.LevelInfo,
			fmt.Sprintf("🟢 DB %s%s%s completed", ColorGreen, operation, ColorReset),
			"duration", duration.String(),
		)
	}
}

func (l *Logger) ExternalCall(ctx context.Context, service, endpoint string, duration time.Duration, err error) {
	if err != nil {
		l.logWithContext(ctx, slog.LevelError,
			fmt.Sprintf("🔴 External call to %s%s%s failed", ColorRed, service, ColorReset),
			"endpoint", endpoint,
			"duration", duration.String(),
			"error", err.Error(),
		)
	} else {
		l.logWithContext(ctx, slog.LevelInfo,
			fmt.Sprintf("🟢 External call to %s%s%s completed", ColorGreen, service, ColorReset),
			"endpoint", endpoint,
			"duration", duration.String(),
		)
	}
}

// Service startup/shutdown logging
func (l *Logger) ServiceStarted(port string, services ...string) {
	l.InfoMsg(fmt.Sprintf("Service started on port %s%s%s", ColorGreen, port, ColorReset))
	if len(services) > 0 {
		l.InfoMsg(fmt.Sprintf("Connected services: %s%s%s", ColorCyan, strings.Join(services, ", "), ColorReset))
	}
}

func (l *Logger) ServiceStopped() {
	l.InfoMsg(fmt.Sprintf("🛑 Service %sstopped gracefully%s", ColorYellow, ColorReset))
}

// Internal helper method
func (l *Logger) logWithContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	// Extract context values
	contextArgs := l.extractContextArgs(ctx)

	// Combine context args with provided args
	allArgs := append(contextArgs, args...)

	l.Logger.Log(ctx, level, msg, allArgs...)
}

func (l *Logger) extractContextArgs(ctx context.Context) []any {
	var args []any

	if requestID := getFromContext(ctx, RequestIDKey); requestID != "" {
		args = append(args, "request_id", requestID)
	}

	if userID := getFromContext(ctx, UserIDKey); userID != "" {
		args = append(args, "user_id", userID)
	}

	if correlationID := getFromContext(ctx, CorrelationIDKey); correlationID != "" {
		args = append(args, "correlation_id", correlationID)
	}

	return args
}

// Context helper functions
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

func GetRequestID(ctx context.Context) string {
	return getFromContext(ctx, RequestIDKey)
}

func GetUserID(ctx context.Context) string {
	return getFromContext(ctx, UserIDKey)
}

func GetCorrelationID(ctx context.Context) string {
	return getFromContext(ctx, CorrelationIDKey)
}

func GetOrCreateRequestID(ctx context.Context) (context.Context, string) {
	if id := GetRequestID(ctx); id != "" {
		return ctx, id
	}
	requestID := generateID()
	return WithRequestID(ctx, requestID), requestID
}

func GetOrCreateCorrelationID(ctx context.Context) (context.Context, string) {
	if id := GetCorrelationID(ctx); id != "" {
		return ctx, id
	}
	correlationID := generateID()
	return WithCorrelationID(ctx, correlationID), correlationID
}

// Utility functions
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getFromContext(ctx context.Context, key ContextKey) string {
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func generateID() string {
	return uuid.New().String()[:8] // Use short ID for readability
}

// Package level convenience functions
func Info(ctx context.Context, msg string, args ...any) {
	Get().Info(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	Get().Warn(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	Get().Error(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	Get().Debug(ctx, msg, args...)
}

func InfoMsg(msg string, args ...any) {
	Get().InfoMsg(msg, args...)
}

func WarnMsg(msg string, args ...any) {
	Get().WarnMsg(msg, args...)
}

func ErrorMsg(msg string, args ...any) {
	Get().ErrorMsg(msg, args...)
}

func DebugMsg(msg string, args ...any) {
	Get().DebugMsg(msg, args...)
}

func HTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	Get().HTTPRequest(ctx, method, path, statusCode, duration)
}

func Database(ctx context.Context, operation string, duration time.Duration, err error) {
	Get().Database(ctx, operation, duration, err)
}

func ExternalCall(ctx context.Context, service, endpoint string, duration time.Duration, err error) {
	Get().ExternalCall(ctx, service, endpoint, duration, err)
}

func ServiceStarted(port string, services ...string) {
	Get().ServiceStarted(port, services...)
}

func ServiceStopped() {
	Get().ServiceStopped()
}
