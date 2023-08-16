package log

import (
	"os"
	"github.com/sirupsen/logrus"
)

type logrusAdapter struct {
	*logrus.Entry
}

func NewLogrusLogger() Logger {

	logger := logrus.New()

	logger.SetReportCaller(true)

	logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: false,
	})

	logger.SetOutput(os.Stdout)

	return &logrusAdapter{logrus.NewEntry(logger)}
}

func (l *logrusAdapter) WithFields(args map[string]any) Logger {
	return &logrusAdapter{l.Entry.WithFields(args)}
}
