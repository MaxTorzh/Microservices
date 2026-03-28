package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(level string) {
    log = logrus.New()
    
    log.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
    
    log.SetOutput(os.Stdout)
    
    lvl, err := logrus.ParseLevel(level)
    if err != nil {
        lvl = logrus.InfoLevel
    }
    log.SetLevel(lvl)
}

func Get() *logrus.Logger {
    if log == nil {
        Init("info")
    }
    return log
}

func Info(args ...interface{}) {
    Get().Info(args...)
}

func Infof(format string, args ...interface{}) {
    Get().Infof(format, args...)
}

func Error(args ...interface{}) {
    Get().Error(args...)
}

func Errorf(format string, args ...interface{}) {
    Get().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
    Get().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
    Get().Fatalf(format, args...)
}

func Debug(args ...interface{}) {
    Get().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
    Get().Debugf(format, args...)
}

func Warn(args ...interface{}) {
    Get().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
    Get().Warnf(format, args...)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
    return Get().WithFields(fields)
}
