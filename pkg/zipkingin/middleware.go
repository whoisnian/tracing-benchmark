// https://pkg.go.dev/github.com/openzipkin/zipkin-go/middleware/http
package zipkingin

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
)

func Middleware(tracer *zipkin.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		spanContext := tracer.Extract(b3.ExtractHTTP(c.Request))

		var spanName string
		if c.FullPath() != "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		} else {
			spanName = fmt.Sprintf("%s unknown route", c.Request.Method)
		}

		sp := tracer.StartSpan(
			spanName,
			zipkin.Kind(model.Server),
			zipkin.Parent(spanContext),
		)
		ctx := zipkin.NewContext(c.Request.Context(), sp)
		c.Request = c.Request.WithContext(ctx)

		if zipkin.IsNoop(sp) {
			c.Next()
			return
		}

		remoteEndpoint, _ := zipkin.NewEndpoint("", c.Request.RemoteAddr)
		sp.SetRemoteEndpoint(remoteEndpoint)

		zipkin.TagHTTPMethod.Set(sp, c.Request.Method)
		zipkin.TagHTTPPath.Set(sp, c.Request.URL.Path)
		if c.Request.ContentLength > 0 {
			zipkin.TagHTTPRequestSize.Set(sp, strconv.FormatInt(c.Request.ContentLength, 10))
		}

		c.Next()

		zipkin.TagHTTPStatusCode.Set(sp, strconv.Itoa(c.Writer.Status()))
		if c.Writer.Status() > 399 {
			zipkin.TagError.Set(sp, strconv.Itoa(c.Writer.Status()))
		}
		zipkin.TagHTTPResponseSize.Set(sp, strconv.Itoa(c.Writer.Size()))
		sp.Finish()
	}
}
