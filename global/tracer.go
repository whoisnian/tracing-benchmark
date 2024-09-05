package global

import (
	"context"

	"go.elastic.co/apm/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var TR Tracer

type Tracer interface {
	Shutdown(context.Context) error
}

func SetupTracer() {
	switch CFG.TraceBackend {
	case "none":
		TR = setupNopTracer()
	case "otlp":
		TR = setupOtlpTracer()
	case "apm":
		TR = setupApmTracer()
	default:
		panic("unknown trace backend: " + CFG.TraceBackend)
	}
}

type nopTracer struct{}

func setupNopTracer() *nopTracer                  { return &nopTracer{} }
func (*nopTracer) Shutdown(context.Context) error { return nil }

type otlpTracer struct {
	provider *sdktrace.TracerProvider
	itracer  trace.Tracer
}

// https://opentelemetry.io/docs/languages/go/instrumentation/#getting-a-tracer
func setupOtlpTracer() *otlpTracer {
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(CFG.TraceOtlpEndpoint))
	if err != nil {
		panic(err)
	}

	rsc, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(AppName),
			semconv.ServiceVersion(Version),
		),
	)
	if err != nil {
		panic(err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(rsc),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return &otlpTracer{provider, otel.GetTracerProvider().Tracer(ModName)}
}

func (tr *otlpTracer) Shutdown(ctx context.Context) error {
	return tr.provider.Shutdown(ctx)
}

type apmTracer struct {
	itracer *apm.Tracer
}

func setupApmTracer() *apmTracer {
	itracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:    AppName,
		ServiceVersion: Version,
	})
	if err != nil {
		panic(err)
	}
	return &apmTracer{itracer}
}

func (tr *apmTracer) Shutdown(ctx context.Context) error {
	tr.itracer.Flush(ctx.Done())
	tr.itracer.Close()
	return nil
}
