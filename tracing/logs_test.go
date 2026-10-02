package tracingClient

import (
	"context"
	"sync"
	"testing"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type recordingProcessor struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (p *recordingProcessor) OnEmit(_ context.Context, r *sdklog.Record) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.records = append(p.records, r.Clone())
	return nil
}
func (p *recordingProcessor) Enabled(context.Context, sdklog.EnabledParameters) bool { return true }
func (p *recordingProcessor) Shutdown(context.Context) error                         { return nil }
func (p *recordingProcessor) ForceFlush(context.Context) error                       { return nil }

func newTestShipper() (*lineShipper, *recordingProcessor) {
	rec := &recordingProcessor{}
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(rec))
	return &lineShipper{logger: lp.Logger("test")}, rec
}

func TestWriterEmitsOneRecordPerLineAndHoldsPartialLines(t *testing.T) {
	s, rec := newTestShipper()
	w := s.writer("stdout")
	w.Write([]byte("first line\nsecond "))
	w.Write([]byte("half\n\n"))
	if len(rec.records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(rec.records))
	}
	if got := rec.records[1].Body().AsString(); got != "second half" {
		t.Fatalf("partial line was not joined, got %q", got)
	}
}

func TestEmitStripsColourCodes(t *testing.T) {
	s, rec := newTestShipper()
	s.emit("stdout", "\x1b[31;1m/app/repo.go:54 \x1b[35;1mrecord not found\x1b[0m\n")
	if got := rec.records[0].Body().AsString(); got != "/app/repo.go:54 record not found" {
		t.Fatalf("ANSI codes left in body: %q", got)
	}
	if rec.records[0].Severity() != otellog.SeverityWarn {
		t.Fatalf("record not found should be WARN, got %v", rec.records[0].Severity())
	}
}

func TestSeverityOf(t *testing.T) {
	cases := map[string]otellog.Severity{
		`{"time":"x","method":"GET","status":200,"error":""}`:  otellog.SeverityInfo,
		`{"time":"x","method":"GET","status":404,"error":""}`:  otellog.SeverityWarn,
		`{"time":"x","method":"GET","status":503,"error":"x"}`: otellog.SeverityError,
		`{"level":"ERROR","msg":"boom"}`:                       otellog.SeverityError,
		`2026/10/02 10:05:32 Error fetching meter balance`:     otellog.SeverityError,
		`2026/10/02 10:05:32 panic recovered in handler`:       otellog.SeverityFatal,
		`2026/10/02 10:05:32 --> GET http://user-service:8080`: otellog.SeverityInfo,
	}
	for line, want := range cases {
		if got, _ := severityOf(line); got != want {
			t.Errorf("%q: got %v, want %v", line, got, want)
		}
	}
}

func TestPulsarFilterDropsChatterAndEventDumpsButKeepsTheRest(t *testing.T) {
	s, rec := newTestShipper()
	f := &pulsarFilter{next: s.writer("stderr")}
	f.Write([]byte("2026/10/02 11:08:54 \x1b[36m[Pulsar] Published event 'x' to topic 'y'\x1b[0m\n"))
	f.Write([]byte("2026/10/02 11:08:54 ============================== Consumer ===============================\n"))
	f.Write([]byte("2026/10/02 11:08:54 EventType: water_credit.budget.index\n"))
	f.Write([]byte("2026/10/02 11:08:54 Payload:\n{\n  \"phoneNumber\": \"+233000000000\"\n}\n"))
	f.Write([]byte("2026/10/02 11:08:54 Pulsar Handler: Received event 'a' from topic 'b'\n"))
	f.Write([]byte("2026/10/02 11:08:55 Error fetching meter balance for 68753500111394\n"))
	if len(rec.records) != 1 {
		t.Fatalf("expected only the service's own line to be shipped, got %d records", len(rec.records))
	}
	if got := rec.records[0].Body().AsString(); got != "2026/10/02 11:08:55 Error fetching meter balance for 68753500111394" {
		t.Fatalf("unexpected record %q", got)
	}
}
