package global

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	skywalkingtrace "github.com/apache/skywalking-go/toolkit/trace"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"go.elastic.co/apm/v2"
	"go.opentelemetry.io/otel/trace"
)

var LOG *slog.Logger

func SetupLogger() {
	LOG = slog.New(&TraceHandler{
		slog.NewJSONHandler(
			os.Stderr,
			&slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo},
		),
	})
}

type TraceHandler struct{ slog.Handler }

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	var (
		traceID       = "unknown"
		spanID        = "unknown"
		transactionID = "unknown"
		segmentID     = "unknown"
	)
	switch CFG.TraceBackend {
	case "otlp":
		// https://github.com/open-telemetry/opentelemetry-go/blob/ed4fc757583a88b4da51b1fe1c3f0703ac27a487/sdk/log/logger.go#L73
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			traceID = sc.TraceID().String()
			spanID = sc.SpanID().String()
		}
		r.AddAttrs(
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
		)
	case "apm":
		// https://github.com/elastic/apm-agent-go/blob/096f5c06b782ae2b7c59d9eb4092a63a9a1886bd/module/apmzap/fields.go#L42
		if tx := apm.TransactionFromContext(ctx); tx != nil {
			traceID = tx.TraceContext().Trace.String()
			transactionID = tx.TraceContext().Span.String()
			if span := apm.SpanFromContext(ctx); span != nil {
				spanID = span.TraceContext().Span.String()
			}
		}
		r.AddAttrs(
			slog.String("trace.id", traceID),
			slog.String("transaction.id", transactionID),
			slog.String("span.id", spanID),
		)
	case "zipkin":
		if sc := opentracing.SpanFromContext(ctx); sc != nil {
			traceID = sc.Context().(zipkinot.SpanContext).TraceID.String()
			spanID = sc.Context().(zipkinot.SpanContext).ID.String()
		}
		r.AddAttrs(
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
		)
	case "skywalking":
		traceID = skywalkingtrace.GetTraceID()
		segmentID = skywalkingtrace.GetSegmentID()
		spanID = strconv.FormatInt(int64(skywalkingtrace.GetSpanID()), 10)
		r.AddAttrs(
			slog.String("traceId", traceID),
			slog.String("segmentId", segmentID),
			slog.String("spanId", spanID),
		)
	}

	return h.Handler.Handle(ctx, r)
}
