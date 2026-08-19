package tracingClient

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// samplerFromEnv defaults to sampling 10% of traces. The TracerProvider had
// no sampler set at all before (AlwaysSample by default), and the GORM OTel
// plugin traces every DB query as its own span — once that landed platform-
// wide, every service started sending far more spans than the shared
// otel-collector can absorb, which trips its memory_limiter processor and
// starts rejecting data ("data refused due to high memory usage", visible in
// every service's logs). ParentBased so a trace that a caller already
// decided to sample stays fully sampled end-to-end. Override with
// OTEL_TRACES_SAMPLE_RATIO (0.0–1.0) per service if needed.
func samplerFromEnv() sdktrace.Sampler {
	ratio := 0.1
	if v := os.Getenv("OTEL_TRACES_SAMPLE_RATIO"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed >= 0 && parsed <= 1 {
			ratio = parsed
		}
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
}

// TracingConfig points at the OTel Collector deployed by
// Grundfos.Flow.V1.DevOps/core/jaeger. Spans flow through otel-collector
// (which derives RED metrics for the Jaeger Monitor/SPM tab via the
// spanmetrics connector) and are then forwarded to jaeger-collector for
// storage. Override via OTEL_EXPORTER_OTLP_ENDPOINT env var.
type TracingConfig struct {
	ServiceName string
	Endpoint    string // e.g. "otel-collector.observability.svc.cluster.local:4317"
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
		cfg.Endpoint = "otel-collector.observability.svc.cluster.local:4317"
	}
	return &tracingClient{config: cfg}
}

// Connect wires up the global OTel TracerProvider and MeterProvider with
// OTLP/gRPC exporters pointed at the shared otel-collector, starts emitting
// Go runtime metrics (goroutines, GC, memory — zero code changes needed in
// callers), and returns a shutdown func to flush both on exit.
func (c *tracingClient) Connect() func(context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(c.config.ServiceName),
		),
	)
	if err != nil {
		log.Println("tracing: failed to build resource", err)
		res = resource.Default()
	}

	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(c.config.Endpoint)}
	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(c.config.Endpoint)}
	if c.config.Insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}

	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		log.Println("tracing: failed to connect trace exporter to OTLP endpoint", err)
		return func(context.Context) error { return nil }
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(samplerFromEnv()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdownFuncs := []func(context.Context) error{tp.Shutdown}

	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		// Non-fatal: traces still work without metrics.
		log.Println("tracing: failed to connect metric exporter to OTLP endpoint", err)
	} else {
		mp := metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(metricExporter)),
			metric.WithResource(res),
		)
		otel.SetMeterProvider(mp)
		shutdownFuncs = append(shutdownFuncs, mp.Shutdown)

		if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
			log.Println("tracing: failed to start Go runtime metrics", err)
		}
	}

	log.Println("tracing: connected to otel-collector at", c.config.Endpoint)
	return func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}
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
