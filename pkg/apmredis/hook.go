// https://pkg.go.dev/go.elastic.co/apm/module/apmgoredisv8/v2
// https://pkg.go.dev/github.com/redis/go-redis/extra/redisotel/v9
package apmredis

import (
	"bytes"
	"context"
	"net"

	"github.com/redis/go-redis/v9"
	"go.elastic.co/apm/v2"
)

type hook struct{}

func NewHook() redis.Hook {
	return &hook{}
}

func (h *hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		span, ctx := apm.StartSpanOptions(ctx, "redis.dial", "db.redis", apm.SpanOptions{ExitSpan: true})
		defer span.End()

		conn, err := next(ctx, network, addr)
		if err != nil {
			apm.CaptureError(ctx, err).Send()
			return nil, err
		}
		return conn, nil
	}
}

func (h *hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		span, ctx := apm.StartSpanOptions(ctx, cmd.FullName(), "db.redis", apm.SpanOptions{ExitSpan: true})
		defer span.End()

		if err := next(ctx, cmd); err != nil {
			apm.CaptureError(ctx, err).Send()
			return err
		}
		return nil
	}
}

func (h *hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		nameBuf := bytes.NewBufferString("redis.pipeline")
		for _, cmd := range cmds {
			nameBuf.Write([]byte(", "))
			nameBuf.WriteString(cmd.FullName())
		}
		span, ctx := apm.StartSpanOptions(ctx, nameBuf.String(), "db.redis", apm.SpanOptions{ExitSpan: true})
		defer span.End()

		if err := next(ctx, cmds); err != nil {
			apm.CaptureError(ctx, err).Send()
			return err
		}
		return nil
	}
}
