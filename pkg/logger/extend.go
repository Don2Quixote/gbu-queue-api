package logger

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

// Extend extends logger with key if logger has supported type.
func Extend(log Logger, key string, value interface{}) Logger {
	switch log := log.(type) {
	case *logrus.Logger:
		return log.WithField(key, value)
	case *logrus.Entry:
		return log.WithField(key, value)
	default:
		log.Warnf("logger is not extendable (%s)", reflect.TypeOf(log).String())
		return log
	}
}
