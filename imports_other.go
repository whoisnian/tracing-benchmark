//go:build !skywalking

package main

import "github.com/whoisnian/tracing-benchmark/global"

func MustMatchTracer() {
	if global.CFG.TraceBackend == "skywalking" {
		panic("can not use skywalking trace backend")
	}
}
