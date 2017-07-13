package logger

import (
	"os"
	"fmt"
	"path"
	"time"
	"errors"
	"runtime"
	"path/filepath"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/natefinch/lumberjack"
)

type Log struct {
}

// Configuration for logging
type Config struct {
	// EncodeLogsAsJson makes the log framework log JSON
	EncodeLogsAsJson bool
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
	// StackStrace make debug log stack
	StackStrace bool
	// LogLevel log level
	LogLevel zapcore.Level
}

// How to log, by example:
// logger.Info("Importing new file, zap.String("source", filename), zap.Int("size", 1024))
// To log a stacktrace:
// logger.Error("It went wrong, zap.Stack())

// DefaultZapLogger is the default logger instance that should be used to log
// It's assigned a default value here for tests (which do not call log.Configure())
var DefaultZapLogger = newZapLogger(false, os.Stdout)
var DefaultLoggerConfig Config

func Bool(name string, value bool) zapcore.Field {
	return zap.Bool(name, value)
}

func Int(name string, value int) zapcore.Field {
	return zap.Int(name, value)
}

func String(name string, value string) zapcore.Field {
	return zap.String(name, value)
}

func Int64(name string, value int64) zapcore.Field {
	return zap.Int64(name, value)
}

func Duration(name string, value time.Duration) zapcore.Field {
	return zap.Duration(name, value)
}

func Err(err error) zapcore.Field {
	return zap.Error(err)
}

// Debug Log a message at the debug level. Messages include any context that's
// accumulated on the logger, as well as any fields added at the log site.
//
// Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (l *Log) Debug(msg string, fields ...zapcore.Field) {
	if DefaultLoggerConfig.StackStrace {
		fields = append(fields, Stack())
		DefaultZapLogger.Debug(msg, fields...)
	} else {
	  DefaultZapLogger.Debug(msg, fields...)
	}
}

// Info log a message at the info level. Messages include any context that's
// accumulated on the logger, as well as any fields added at the log site.
//
// Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (l *Log) Info(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Info(msg, fields...)
}

// Warn log a message at the warn level. Messages include any context that's
// accumulated on the logger, as well as any fields added at the log site.
//
// Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (l *Log) Warn(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Warn(msg, fields...)
}

// Error Log a message at the error level. Messages include any context that's
// accumulated on the logger, as well as any fields added at the log site.
//
// Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (l *Log) Error(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Error(msg, fields...)
}

// Panic Log a message at the Panic level. Messages include any context that's
// accumulated on the logger, as well as any fields added at the log site.
//
// Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (l *Log) Panic(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Panic(msg, fields...)
}

// Fatal Log a message at the fatal level. Messages include any context that's
// accumulated on the logger, as well as any fields added at the log site.
//
// Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (l *Log) Fatal(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Fatal(msg, fields...)
}

func Stack() zapcore.Field {
	pc, file, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%s:%d", file, runtime.FuncForPC(pc).Name(), lineno)
	}
	return zap.String("stacktrace", src)
}

//// AtLevel logs the message at a specific log level
//func AtLevel(level zapcore.Level, msg string, fields ...zapcore.Field) {
//	switch level {
//	case zapcore.DebugLevel:
//		Debug(msg, fields...)
//	case zapcore.PanicLevel:
//		Panic(msg, fields...)
//	case zapcore.ErrorLevel:
//		Error(msg, fields...)
//	case zapcore.WarnLevel:
//		Warn(msg, fields...)
//	case zapcore.InfoLevel:
//		Info(msg, fields...)
//	case zapcore.FatalLevel:
//		Fatal(msg, fields...)
//	default:
//		Warn("Logging at unkown level", zap.Any("level", level))
//		Warn(msg, fields...)
//	}
//}

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/auth-service/auth-service.log and
// will be rolled when it reaches 20MB with a maximum of 1 backup.
func Configure(config Config) {
	writers := []zapcore.WriteSyncer{os.Stdout}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}

	DefaultZapLogger = newZapLogger(config.EncodeLogsAsJson, zapcore.NewMultiWriteSyncer(writers...))
	zap.RedirectStdLog(DefaultZapLogger)
	//Info("logging configured",
	//	zap.Bool("fileLogging", config.FileLoggingEnabled),
	//	zap.Bool("jsonLogOutput", config.EncodeLogsAsJson),
	//	zap.String("logDirectory", config.Directory),
	//	zap.String("fileName", config.Filename),
	//	zap.Int("maxSizeMB", config.MaxSize),
	//	zap.Int("maxBackups", config.MaxBackups),
	//	zap.Int("maxAgeInDays", config.MaxAge))
	DefaultLoggerConfig = config
}

func Init(file, level string, size, backup int, stackstrace bool) (Log, error) {
	log := Log{}

	name := filepath.Base(file)
	if name == "" {
		return log, errors.New("Bad file")
	}
	dir := filepath.Dir(file)
	if dir == "" {
		dir = "./"
	}
	if size < 0 || backup < 0 {
		return log, errors.New("Bad size or backup")
	}

	config := Config{
		EncodeLogsAsJson: true,
		FileLoggingEnabled: true,
		Directory: dir,
		Filename: name,
		MaxSize: size,
		MaxBackups: backup,
		StackStrace: stackstrace,
	}

	if level == "" {
		level = "error"
	}
	if err := SetLogLevel(level); err != nil {
		return log, err
	}

	Configure(config)

	return log, nil
}

func newRollingFile(config Config) zapcore.WriteSyncer {
	if err := os.MkdirAll(config.Directory, 0); err != nil {
		fmt.Printf("Failed create log directory in %s, error: %s\n", config.Directory, err)
		return nil
	}

	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxSize:    config.MaxSize,    //megabytes
		MaxAge:     config.MaxAge,     //days
		MaxBackups: config.MaxBackups, //files
	})
}

func newZapLogger(encodeAsJSON bool, output zapcore.WriteSyncer) *zap.Logger {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		//EncodeTime:     zapcore.EpochNanosTimeEncoder,
		//EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeTime:     zapcore.EpochMillisTimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encCfg)
	if encodeAsJSON {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	return zap.New(zapcore.NewCore(encoder, output, zap.NewAtomicLevelAt(DefaultLoggerConfig.LogLevel)))
}

func SetLogLevel(level string) error {
	if level == "debug" {
		DefaultLoggerConfig.LogLevel = zap.DebugLevel
	} else if level == "info" {
		DefaultLoggerConfig.LogLevel = zap.InfoLevel
	} else if level == "warn" {
		DefaultLoggerConfig.LogLevel = zap.WarnLevel
	} else if level == "error" {
		DefaultLoggerConfig.LogLevel = zap.ErrorLevel
	} else {
		return errors.New("Bad log level")
	}

	return nil
}
