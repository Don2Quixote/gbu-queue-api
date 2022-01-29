package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// NewLogrus returns setted up logrus logger.
func NewLogrus() *logrus.Entry {
	l := &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.DebugLevel,
		Formatter: &logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		},
	}
	return l.WithTime(time.Now())
}
