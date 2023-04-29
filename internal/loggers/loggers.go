package loggers

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Conf struct {
	FileName    string
	TimeFormat  string
	MaxSizeMb   int
	MaxBackups  int
	MaxAgeDays  int
	MultiWriter bool
	Level       logrus.Level
}

func (conf Conf) CreateLoggerWithRotate(fileName string) *logrus.Logger {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    conf.MaxSizeMb,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAgeDays,
		Compress:   true,
	}

	logger := &logrus.Logger{
		Out:   lumberjackLogger,
		Level: conf.Level,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: conf.TimeFormat,
		},
	}

	if conf.MultiWriter {
		logger.Out = io.MultiWriter(os.Stdout, lumberjackLogger)
	}

	return logger
}
