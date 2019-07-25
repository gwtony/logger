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

type Config struct {
	EncodeLogsAsJson bool
	FileLoggingEnabled bool
	Directory string
	Filename string
	MaxSize int
	MaxBackups int
	MaxAge int
	StackStrace bool

	LogLevel zapcore.Level
}

var DefaultZapLogger *zap.Logger
var DefaultZapSugaredLogger *zap.SugaredLogger
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

func Error(err error) zapcore.Field {
	return zap.Error(err)
}

func (l *Log) Debug(msg string, fields ...zapcore.Field) {
	if DefaultLoggerConfig.StackStrace {
		fields = append(fields, Stack())
		DefaultZapLogger.Debug(msg, fields...)
	} else {
	  DefaultZapLogger.Debug(msg, fields...)
	}
}

// *f prints "msg": "xxxx"
// * prints "msg": "xxx", "key": "value"
func (l *Log) Info(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Info(msg, fields...)
}
func (l *Log) Infof(msg string, fields ...interface{}) {
	DefaultZapSugaredLogger.Infof(msg, fields...)
}

func (l *Log) Warn(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Warn(msg, fields...)
}
func (l *Log) Warnf(msg string, fields ...interface{}) {
	DefaultZapSugaredLogger.Warnf(msg, fields...)
}

func (l *Log) Error(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Error(msg, fields...)
}
func (l *Log) Errorf(msg string, fields ...interface{}) {
	DefaultZapSugaredLogger.Errorf(msg, fields...)
}

func (l *Log) Panic(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Panic(msg, fields...)
}
func (l *Log) Panicf(msg string, fields ...interface{}) {
	DefaultZapSugaredLogger.Panicf(msg, fields...)
}

func (l *Log) Fatal(msg string, fields ...zapcore.Field) {
	DefaultZapLogger.Fatal(msg, fields...)
}
func (l *Log) Fatalf(msg string, fields ...interface{}) {
	DefaultZapSugaredLogger.Fatalf(msg, fields...)
}

func Stack() zapcore.Field {
	pc, file, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%s:%d", file, runtime.FuncForPC(pc).Name(), lineno)
	}
	return zap.String("stacktrace", src)
}

func Configure(config Config) {
	writers := []zapcore.WriteSyncer{os.Stdout}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}

	DefaultZapLogger = newZapLogger(config.EncodeLogsAsJson, zapcore.NewMultiWriteSyncer(writers...))
	zap.RedirectStdLog(DefaultZapLogger)
	DefaultZapSugaredLogger = DefaultZapLogger.Sugar()
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
		LocalTime: true,
	})
}

func newZapLogger(encodeAsJSON bool, output zapcore.WriteSyncer) *zap.Logger {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder, //default print time in "2019-07-25T16:49:46.037+0800"
		//EncodeTime:     zapcore.EpochMillisTimeEncoder,
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
