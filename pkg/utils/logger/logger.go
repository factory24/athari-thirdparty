package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	// logrus.SetLevel(logrus.WarnLevel)
}

// NewLogger creates a new logger instance with optional fields
func NewLogger(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}

// SetLevel sets the global log level
func SetLevel(level string) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
		return
	}
	logrus.SetLevel(l)
}
