package anet

import (
	"github.com/sirupsen/logrus"
)

func newDefaultLogger() *logrus.Logger {
	return logrus.New()
}

func setLogger(logger *logrus.Logger) {
	log = logger
}

func setLoggerLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

var log *logrus.Logger
