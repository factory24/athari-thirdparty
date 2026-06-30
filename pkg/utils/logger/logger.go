package logger

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Default level can be overridden with LOG_LEVEL env var (see SetLevelFromEnv).
	logrus.SetLevel(logrus.InfoLevel)
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

// SetLevelFromEnv reads LOG_LEVEL (debug|info|warn|error) and applies it,
// defaulting to info when unset or invalid. Call once at service startup.
func SetLevelFromEnv() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		logrus.SetLevel(logrus.InfoLevel)
		return
	}
	SetLevel(level)
}

// FromEchoContext returns a logger entry pre-populated with the trace_id set
// by tracingClient.EchoMiddleware (ThirdParty/tracing), so every log line for
// a request can be correlated with its Jaeger trace. Falls back to a plain
// entry if no trace_id is present (e.g. tracing middleware not registered).
func FromEchoContext(c echo.Context, fields logrus.Fields) *logrus.Entry {
	if fields == nil {
		fields = logrus.Fields{}
	}
	if traceID, ok := c.Get("trace_id").(string); ok && traceID != "" {
		fields["trace_id"] = traceID
	}
	fields["path"] = c.Path()
	return logrus.WithFields(fields)
}
