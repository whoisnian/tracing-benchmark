//go:build skywalking

package main

import (
	_ "github.com/apache/skywalking-go"
	"github.com/whoisnian/tracing-benchmark/global"
)

func MustMatchTracer() {
	if global.CFG.TraceBackend != "skywalking" {
		panic("only skywalking trace backend can be used")
	}
}
