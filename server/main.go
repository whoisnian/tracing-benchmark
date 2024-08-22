package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/whoisnian/tracing-benchmark/server/global"
)

func main() {
	global.SetupConfig()
	if global.CFG.Version {
		fmt.Printf("%s %s(%s)\n", global.AppName, global.Version, global.BuildTime)
		return
	}

	go func() {
		if err := http.ListenAndServe(global.CFG.ListenAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	waitFor(syscall.SIGINT, syscall.SIGTERM)
}

func waitFor(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	defer signal.Stop(c)

	<-c
}
