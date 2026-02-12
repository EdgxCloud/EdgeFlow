package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// BroadcastFunc is called for each log entry to send to WebSocket clients
type BroadcastFunc func(level, message, source string, fields map[string]interface{})

var (
	globalLogger *zap.Logger
	globalSugar  *zap.SugaredLogger
	broadcastFn  BroadcastFunc
	mu           sync.RWMutex
)

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	LogDir     string // directory for log files (empty = no file logging)
	MaxSizeMB  int    // max size per log file in MB
	MaxBackups int    // max number of old log files
	MaxAgeDays int    // max days to retain old log files
	Compress   bool   // gzip compress rotated files
}

// DefaultConfig returns sensible defaults for Raspberry Pi
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "console",
		LogDir:     "./logs",
		MaxSizeMB:  50,
		MaxBackups: 5,
		MaxAgeDays: 7,
		Compress:   true,
	}
}

// Init initializes the global logger with the given configuration
func Init(cfg Config) error {
	logLevel, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	// 1. Console output (always on)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel))

	// 2. JSON file with rotation (if logDir set)
	if cfg.LogDir != "" {
		if mkErr := os.MkdirAll(cfg.LogDir, 0755); mkErr != nil {
			return fmt.Errorf("failed to create log directory: %w", mkErr)
		}
		fileWriter := &lumberjack.Logger{
			Filename:   filepath.Join(cfg.LogDir, "edgeflow.log"),
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		}
		jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)
		cores = append(cores, zapcore.NewCore(jsonEncoder, zapcore.AddSync(fileWriter), logLevel))
	}

	// 3. WebSocket bridge (broadcasts to frontend LogPanel)
	cores = append(cores, &wsBridgeCore{level: logLevel})

	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))

	mu.Lock()
	globalLogger = logger
	globalSugar = logger.Sugar()
	mu.Unlock()

	return nil
}

// SetBroadcaster sets the WebSocket broadcast function.
// Called after WebSocket hub is initialized.
func SetBroadcaster(fn BroadcastFunc) {
	mu.Lock()
	defer mu.Unlock()
	broadcastFn = fn
}

// Get returns the global zap.Logger
func Get() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if globalLogger == nil {
		l, _ := zap.NewDevelopment()
		return l
	}
	return globalLogger
}

// Sugar returns the global sugared logger
func Sugar() *zap.SugaredLogger {
	mu.RLock()
	defer mu.RUnlock()
	if globalSugar == nil {
		l, _ := zap.NewDevelopment()
		return l.Sugar()
	}
	return globalSugar
}

// Sync flushes buffered log entries
func Sync() error {
	mu.RLock()
	defer mu.RUnlock()
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// --- Convenience functions ---

func Info(msg string, fields ...zap.Field)  { Get().Info(msg, fields...) }
func Error(msg string, fields ...zap.Field) { Get().Error(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { Get().Warn(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { Get().Debug(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { Get().Fatal(msg, fields...) }

// --- Context loggers ---

// WithFlow returns a logger with flow context
func WithFlow(flowID, flowName string) *zap.Logger {
	return Get().With(zap.String("flow_id", flowID), zap.String("flow_name", flowName))
}

// WithNode returns a logger with node context
func WithNode(nodeID, nodeType string) *zap.Logger {
	return Get().With(zap.String("node_id", nodeID), zap.String("node_type", nodeType))
}

// WithFlowNode returns a logger with both flow and node context
func WithFlowNode(flowID, nodeID, nodeType string) *zap.Logger {
	return Get().With(
		zap.String("flow_id", flowID),
		zap.String("node_id", nodeID),
		zap.String("node_type", nodeType),
	)
}

// --- io.Writer adapter for stdlib log compatibility ---

// Writer returns an io.Writer that writes to the logger at Info level.
// Use with: log.SetOutput(logger.Writer())
func Writer() io.Writer {
	return &logWriter{}
}

type logWriter struct{}

func (w *logWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	Get().Info(msg)
	return len(p), nil
}

// --- WebSocket bridge zapcore.Core ---

type wsBridgeCore struct {
	level  zapcore.Level
	fields []zapcore.Field
}

func (c *wsBridgeCore) Enabled(lvl zapcore.Level) bool {
	return lvl >= c.level
}

func (c *wsBridgeCore) With(fields []zapcore.Field) zapcore.Core {
	combined := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	combined = append(combined, c.fields...)
	combined = append(combined, fields...)
	return &wsBridgeCore{level: c.level, fields: combined}
}

func (c *wsBridgeCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		ce = ce.AddCore(entry, c)
	}
	return ce
}

func (c *wsBridgeCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	mu.RLock()
	fn := broadcastFn
	mu.RUnlock()
	if fn == nil {
		return nil
	}

	level := "info"
	switch entry.Level {
	case zapcore.DebugLevel:
		level = "debug"
	case zapcore.WarnLevel:
		level = "warn"
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		level = "error"
	}

	source := "backend"
	extra := make(map[string]interface{})

	// Process both core-level fields and entry-level fields
	allFields := append(c.fields, fields...)
	for _, f := range allFields {
		switch f.Key {
		case "source":
			source = f.String
		default:
			switch f.Type {
			case zapcore.StringType:
				extra[f.Key] = f.String
			case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
				extra[f.Key] = f.Integer
			case zapcore.Float64Type:
				extra[f.Key] = float64(f.Integer)
			case zapcore.BoolType:
				extra[f.Key] = f.Integer == 1
			case zapcore.DurationType:
				extra[f.Key] = time.Duration(f.Integer).String()
			case zapcore.ErrorType:
				if f.Interface != nil {
					extra[f.Key] = fmt.Sprintf("%v", f.Interface)
				}
			}
		}
	}

	fn(level, entry.Message, source, extra)
	return nil
}

func (c *wsBridgeCore) Sync() error { return nil }
