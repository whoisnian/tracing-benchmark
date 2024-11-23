// https://pkg.go.dev/github.com/opentracing-contrib/goredis
// https://pkg.go.dev/github.com/globocom/go-redis-opentracing
// https://pkg.go.dev/github.com/redis/go-redis/extra/redisotel/v9
package zipkinredis

import (
	"bytes"
	"context"
	"fmt"
	"net"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/redis/go-redis/v9"
)

type hook struct{}

func NewHook() redis.Hook {
	return &hook{}
}

func (h *hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "redis.dial")
		ext.DBType.Set(span, "redis")
		ext.DBStatement.Set(span, fmt.Sprintf("dial %s %s", network, addr))
		defer span.Finish()

		conn, err := next(ctx, network, addr)
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("db.error", err.Error())
			return nil, err
		}
		return conn, nil
	}
}

func (h *hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		span, ctx := opentracing.StartSpanFromContext(ctx, "redis.cmd")
		ext.DBType.Set(span, "redis")
		ext.DBStatement.Set(span, cmd.FullName())
		defer span.Finish()

		if err := next(ctx, cmd); err != nil {
			ext.Error.Set(span, true)
			span.SetTag("db.error", err.Error())
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
		span, ctx := opentracing.StartSpanFromContext(ctx, "redis.pipeline")
		ext.DBType.Set(span, "redis")
		ext.DBStatement.Set(span, nameBuf.String())
		defer span.Finish()

		if err := next(ctx, cmds); err != nil {
			ext.Error.Set(span, true)
			span.SetTag("db.error", err.Error())
			return err
		}
		return nil
	}
}
