package example

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var singleton *zap.SugaredLogger
var once_logger sync.Once

// InitLogger initializes a thread-safe singleton logger
// This would be called from a main method when the application starts up
func InitLogger(logPath string, logLevel zapcore.Level, enableStdOut bool) {
	// once ensures the singleton is initialized only once
	once_logger.Do(func() {
		singleton = newLogger(logPath, logLevel, enableStdOut).Named("VMagent")
	})
}

// InitStdOutLogger initialize a stdout logger, just for UT
func InitStdOutLogger(logLevel zapcore.Level) {
	once_logger.Do(func() {
		singleton = newStdOutLogger(logLevel).Named("VMAgent")
	})
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("Jan  2 15:04:05"))
}

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

// GetLogger get a named logger, if name is empty, use the default logger
func GetLogger(name string) (*zap.SugaredLogger, error) {
	if singleton == nil {
		return nil, fmt.Errorf("empty logger: InitLogger method should be called before calling GetLogger")
	}
	if name == "" {
		return singleton, nil
	}
	return singleton.Named(name), nil
}

// GetDefaultLogger get a default logger
func GetDefaultLogger() (*zap.SugaredLogger, error) {
	return GetLogger("")
}

// newLogger create a root logger with provided log path and log level
func newLogger(logPath string, level zapcore.Level, enableStdOut bool) *zap.SugaredLogger {
	fileSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // megabytes
		MaxBackups: 10,  // max file number
		MaxAge:     7,   // days
	})

	writeSyncer := fileSyncer
	if enableStdOut {
		stdoutSyncer := zapcore.AddSync(os.Stdout)
		writeSyncer = zapcore.NewMultiWriteSyncer(fileSyncer, stdoutSyncer)
	}
	zapConf := zap.NewProductionEncoderConfig()
	zapConf.EncodeTime = SyslogTimeEncoder
	zapConf.EncodeLevel = CustomLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConf),
		writeSyncer,
		level,
	)

	return zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	).Sugar()
}

// newStdOutLogger create a stdout logger with provided log level
func newStdOutLogger(level zapcore.Level) *zap.SugaredLogger {
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	zapConf := zap.NewProductionEncoderConfig()
	zapConf.EncodeTime = SyslogTimeEncoder
	zapConf.EncodeLevel = CustomLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConf),
		stdoutSyncer,
		level,
	)

	return zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	).Sugar()
}
