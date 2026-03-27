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

func Info(args ...any) {
	Get().Info(args...)
}

func Infof(format string, args ...any) {
	Get().Infof(format, args...)
}

func Error(args ...any) {
	Get().Error(args...)
}

func Errorf(format string, args ...any) {
	Get().Errorf(format, args...)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}
