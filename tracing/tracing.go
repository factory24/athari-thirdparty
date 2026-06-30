package tracingClient

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig points at the Jaeger OTLP collector deployed by
// Grundfos.Flow.V1.DevOps/core/jaeger. In-cluster default is the
// all-in-one service on port 4317 (gRPC).
type TracingConfig struct {
	ServiceName string
	Endpoint    string // e.g. "jaeger-all-in-one.observability.svc.cluster.local:4317"
	Insecure    bool
}

type TracingClient interface {
	Connect() func(context.Context) error
}

type tracingClient struct {
	config TracingConfig
}

func NewTracingClient(cfg TracingConfig) TracingClient {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "jaeger-all-in-one.observability.svc.cluster.local:4317"
	}
	return &tracingClient{config: cfg}
}

// Connect wires up the global OTel TracerProvider with an OTLP/gRPC
// exporter pointed at Jaeger, and returns a shutdown func to flush on exit.
func (c *tracingClient) Connect() func(context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(c.config.Endpoint)}
	if c.config.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		log.Println("tracing: failed to connect to Jaeger OTLP endpoint", err)
		return func(context.Context) error { return nil }
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(c.config.ServiceName),
		),
	)
	if err != nil {
		log.Println("tracing: failed to build resource", err)
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Println("tracing: connected to Jaeger at", c.config.Endpoint)
	return tp.Shutdown
}

// EchoMiddleware starts a span per inbound request, propagates trace context
// from upstream services (so calls across microservices link into one trace),
// and stashes the trace ID on the echo.Context for log correlation.
func EchoMiddleware(serviceName string) echo.MiddlewareFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := propagator.Extract(c.Request().Context(), propagation.HeaderCarrier(c.Request().Header))

			ctx, span := tracer.Start(ctx, c.Request().Method+" "+c.Path(),
				trace.WithAttributes(
					attribute.String("http.method", c.Request().Method),
					attribute.String("http.route", c.Path()),
				),
			)
			defer span.End()

			c.Set("trace_id", span.SpanContext().TraceID().String())
			c.SetRequest(c.Request().WithContext(ctx))

			err := next(c)
			if err != nil {
				span.RecordError(err)
			}
			span.SetAttributes(attribute.Int("http.status_code", c.Response().Status))
			return err
		}
	}
}

// PropagateOutgoing injects the active trace context into an outgoing
// request's headers so downstream services join the same trace. Call this
// from internal HTTP clients (clients.* packages) before dispatching.
func PropagateOutgoing(ctx context.Context, headers map[string][]string) {
	carrier := propagation.HeaderCarrier(headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}
