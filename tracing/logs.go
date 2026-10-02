package tracingClient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
)

// maxLogLine bounds one record so a dumped payload cannot fill a batch on its own.
const maxLogLine = 16 << 10

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// startLogShipping exports the service's own log output (the standard log package and,
// on Linux, everything written to stdout) over the same OTLP connection as traces.
// Set OTEL_LOGS_EXPORTER=none to turn it off.
func startLogShipping(ctx context.Context, cfg TracingConfig, res *resource.Resource) func(context.Context) error {
	if strings.EqualFold(os.Getenv("OTEL_LOGS_EXPORTER"), "none") {
		return nil
	}
	opts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	} else {
		// Explicit TLS: this exporter reads a scheme-less "host:443" in OTEL_EXPORTER_OTLP_ENDPOINT
		// as a URL whose scheme is the hostname, and falls back to plaintext.
		opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(nil)))
	}
	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		log.Println("tracing: failed to connect log exporter to OTLP endpoint", err)
		return nil
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	shipper := &lineShipper{logger: lp.Logger("github.com/factory24/athari-thirdparty/tracing")}
	log.SetOutput(io.MultiWriter(log.Writer(), shipper.writer("stderr")))
	teeStdout(shipper)
	return lp.Shutdown
}

type lineShipper struct {
	logger otellog.Logger
}

func (s *lineShipper) emit(stream, line string) {
	line = strings.TrimRight(ansiEscape.ReplaceAllString(line, ""), "\r\n\t ")
	if strings.TrimSpace(line) == "" {
		return
	}
	if len(line) > maxLogLine {
		line = line[:maxLogLine]
	}
	now := time.Now()
	var r otellog.Record
	r.SetTimestamp(now)
	r.SetObservedTimestamp(now)
	severity, text := severityOf(line)
	r.SetSeverity(severity)
	r.SetSeverityText(text)
	r.SetBody(otellog.StringValue(line))
	r.AddAttributes(otellog.String("log.iostream", stream))
	s.logger.Emit(context.Background(), r)
}

// writer turns a byte stream into one record per line, holding back a partial line.
func (s *lineShipper) writer(stream string) io.Writer {
	return &lineWriter{emit: func(line string) { s.emit(stream, line) }}
}

type lineWriter struct {
	mu      sync.Mutex
	pending []byte
	emit    func(string)
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending = append(w.pending, p...)
	for {
		i := bytes.IndexByte(w.pending, '\n')
		if i < 0 {
			break
		}
		w.emit(string(w.pending[:i]))
		w.pending = w.pending[i+1:]
	}
	if len(w.pending) > maxLogLine {
		w.emit(string(w.pending))
		w.pending = w.pending[:0]
	}
	return len(p), nil
}

// severityOf reads a JSON line's own level or HTTP status, and otherwise guesses from keywords.
func severityOf(line string) (otellog.Severity, string) {
	if i := strings.IndexByte(line, '{'); i >= 0 && i < 40 && strings.HasSuffix(line, "}") {
		var fields map[string]any
		if json.Unmarshal([]byte(line[i:]), &fields) == nil {
			if lvl, ok := fields["level"].(string); ok {
				return levelSeverity(lvl)
			}
			if status, ok := fields["status"].(float64); ok {
				switch {
				case status >= 500:
					return otellog.SeverityError, "ERROR"
				case status >= 400:
					return otellog.SeverityWarn, "WARN"
				}
				return otellog.SeverityInfo, "INFO"
			}
		}
	}
	lower := strings.ToLower(line)
	for _, k := range []string{"panic", "fatal"} {
		if strings.Contains(lower, k) {
			return otellog.SeverityFatal, "FATAL"
		}
	}
	for _, k := range []string{"error", "failed", "exception", "refused", "timeout"} {
		if strings.Contains(lower, k) {
			return otellog.SeverityError, "ERROR"
		}
	}
	for _, k := range []string{"warn", "record not found"} {
		if strings.Contains(lower, k) {
			return otellog.SeverityWarn, "WARN"
		}
	}
	return otellog.SeverityInfo, "INFO"
}

func levelSeverity(level string) (otellog.Severity, string) {
	switch strings.ToUpper(level) {
	case "DEBUG", "TRACE":
		return otellog.SeverityDebug, "DEBUG"
	case "WARN", "WARNING":
		return otellog.SeverityWarn, "WARN"
	case "ERROR":
		return otellog.SeverityError, "ERROR"
	case "FATAL", "PANIC":
		return otellog.SeverityFatal, "FATAL"
	}
	return otellog.SeverityInfo, "INFO"
}
