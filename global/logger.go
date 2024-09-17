package global

import (
	"context"
	"log/slog"
	"os"

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
		traceID = "unknown"
		spanID  = "unknown"
	)
	switch CFG.TraceBackend {
	case "otlp":
		if sc := trace.SpanFromContext(ctx).SpanContext(); sc.IsValid() {
			traceID = sc.TraceID().String()
			spanID = sc.SpanID().String()
		}
	case "apm":
		if tc := apm.SpanFromContext(ctx).TraceContext(); tc.Trace.Validate() == nil && tc.Span.Validate() == nil {
			traceID = tc.Trace.String()
			spanID = tc.Span.String()
		}
	}

	if CFG.TraceBackend != "none" {
		r.AddAttrs(
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
		)
	}

	return h.Handler.Handle(ctx, r)
}
