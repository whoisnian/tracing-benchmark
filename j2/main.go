package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/stdr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var CFG struct {
	Debug    bool
	Workers  int
	Traces   int
	Service  string
	Endpoint string
}

func init() {
	flag.BoolVar(&CFG.Debug, "debug", false, "Enable debug output")
	flag.IntVar(&CFG.Workers, "workers", 1, "Number of workers (goroutines)")
	flag.IntVar(&CFG.Traces, "traces", 1, "Number of traces for each worker")
	flag.StringVar(&CFG.Service, "service", "oteltrace", "Service name")
	flag.StringVar(&CFG.Endpoint, "endpoint", "http://127.0.0.1:4318", "OTLP http trace exporter endpoint")
	flag.Parse()
}

type countExporter struct {
	*otlptrace.Exporter
	Sum *atomic.Int64
}

func (e *countExporter) ExportSpans(ctx context.Context, ss []sdktrace.ReadOnlySpan) error {
	e.Sum.Add(int64(len(ss)))
	return e.Exporter.ExportSpans(ctx, ss)
}
func (e *countExporter) Shutdown(ctx context.Context) error {
	log.Printf("shutdown countExporter %d", e.Sum.Load())
	return e.Exporter.Shutdown(ctx)
}

type countSpanProcessor struct {
	Sum *atomic.Int64
}

func (sp *countSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {}
func (sp *countSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan)                            { sp.Sum.Add(1) }
func (sp *countSpanProcessor) Shutdown(ctx context.Context) error {
	log.Printf("shutdown countSpanProcessor %d", sp.Sum.Load())
	return nil
}
func (sp *countSpanProcessor) ForceFlush(ctx context.Context) error { return nil }

func main() {
	if CFG.Debug {
		stdr.SetVerbosity(8)
	}

	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(CFG.Endpoint))
	panicIf(err)

	rsc, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(CFG.Service),
		),
	)
	panicIf(err)

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(&countExporter{exporter, new(atomic.Int64)}),
		sdktrace.WithSpanProcessor(&countSpanProcessor{new(atomic.Int64)}),
		sdktrace.WithResource(rsc),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	wg := new(sync.WaitGroup)
	wg.Add(CFG.Workers)
	for i := 0; i < CFG.Workers; i++ {
		go handle(wg, i)
	}
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	panicIf(provider.Shutdown(ctx))
}

func handle(wg *sync.WaitGroup, index int) {
	defer wg.Done()
	log.Printf("start  worker %d", index)

	for i := 0; i < CFG.Traces; i++ {
		// parent span
		ctx, span := otel.GetTracerProvider().Tracer(
			"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin",
			oteltrace.WithInstrumentationVersion("0.56.0"),
		).Start(context.Background(), "/ping/GRM", oteltrace.WithSpanKind(oteltrace.SpanKindServer))
		span.SetAttributes(ginAttributes...)

		// child span (redis)
		_, redisSpan := otel.GetTracerProvider().Tracer(
			"github.com/redis/go-redis/extra/redisotel",
			oteltrace.WithInstrumentationVersion("semver:9.7.0"),
		).Start(ctx, "ping", oteltrace.WithSpanKind(oteltrace.SpanKindClient))
		redisSpan.SetAttributes(redisAttributes...)
		time.Sleep(time.Microsecond * 100)
		redisSpan.End()

		// child span (mysql)
		_, mysqlSpan := otel.GetTracerProvider().Tracer(
			"gorm.io/plugin/opentelemetry",
			oteltrace.WithInstrumentationVersion("0.1.8"),
		).Start(ctx, "gorm.Raw", oteltrace.WithSpanKind(oteltrace.SpanKindClient))
		mysqlSpan.SetAttributes(mysqlAttributes...)
		time.Sleep(time.Microsecond * 100)
		mysqlSpan.End()

		span.End()
	}
	log.Printf("finish worker %d", index)
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	ginAttributes = []attribute.KeyValue{
		semconv.HTTPRequestMethodGet,
		semconv.HTTPRoute("/ping/GRM"),
		attribute.String("http.scheme", "http"),          // exists in go.opentelemetry.io/otel/semconv/v1.20.0
		attribute.Int("http.status_code", http.StatusOK), // exists in go.opentelemetry.io/otel/semconv/v1.20.0
		attribute.String("http.target", "/ping/GRM"),     // exists in go.opentelemetry.io/otel/semconv/v1.20.0
		semconv.NetworkTypeIpv4,
		semconv.NetworkLocalAddress("172.18.0.4"),
		semconv.NetworkLocalPort(8080),
		semconv.NetworkProtocolName("http"),
		semconv.NetworkProtocolVersion("1.1"),
		semconv.NetworkPeerAddress("172.18.0.1"),
		semconv.NetworkPeerPort(58868),
	}
	redisAttributes = []attribute.KeyValue{
		semconv.CodeFilepath("github.com/whoisnian/tracing-benchmark/router/handler.go"),
		semconv.CodeFunction("router.pingRedis"),
		semconv.CodeLineNumber(49),
		semconv.DBSystemRedis,
		attribute.String("db.connection_string", "redis://redis:6379"), // exists in go.opentelemetry.io/otel/semconv/v1.24.0
		semconv.ServerAddress("redis"),
		semconv.ServerPort(6379),
		semconv.DBOperationName("ping"),
		semconv.DBQueryText("ping"),
	}
	mysqlAttributes = []attribute.KeyValue{
		semconv.DBSystemMySQL,
		semconv.ServerAddress("mysql"),
		semconv.ServerPort(3306),
		semconv.DBOperationName("select"),
		semconv.DBQueryText("SELECT 1"),
		attribute.Int("db.rows_affected", 0), // exists in gorm.io/plugin/opentelemetry@v0.1.8
	}
)
