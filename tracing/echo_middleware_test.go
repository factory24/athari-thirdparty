package tracingClient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func serve(t *testing.T, status int) sdktrace.ReadOnlySpan {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	e := echo.New()
	e.Use(EchoMiddleware("svc"))
	e.GET("/x", func(c echo.Context) error { return c.NoContent(status) })
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	return spans[0]
}

func TestEchoMiddlewareMarksInboundRequestsAsServerSpans(t *testing.T) {
	if k := serve(t, http.StatusOK).SpanKind(); k != trace.SpanKindServer {
		t.Fatalf("expected a server span, got %v", k)
	}
}

func TestEchoMiddlewareFlagsOnlyServerErrors(t *testing.T) {
	if s := serve(t, http.StatusInternalServerError).Status().Code; s != codes.Error {
		t.Fatalf("a 500 must mark the span as an error, got %v", s)
	}
	if s := serve(t, http.StatusNotFound).Status().Code; s == codes.Error {
		t.Fatal("a 404 is the caller's mistake and must not mark the span as an error")
	}
}
