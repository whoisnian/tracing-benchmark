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
	global.SetupLogger()
	global.LOG.Info(fmt.Sprintf("setup config successfully: %+v", global.CFG))

	if global.CFG.Version {
		fmt.Printf("%s %s(%s)\n", global.AppName, global.Version, global.BuildTime)
		return
	}

	go func() {
		global.LOG.Info("service is starting: " + global.CFG.ListenAddr)
		if err := http.ListenAndServe(global.CFG.ListenAddr, nil); errors.Is(err, http.ErrServerClosed) {
			global.LOG.Warn("service is shutting down")
		} else {
			global.LOG.Error(err.Error())
			os.Exit(1)
		}
	}()

	waitFor(syscall.SIGINT, syscall.SIGTERM)

	global.LOG.Info("service has been shut down")
}

func waitFor(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	defer signal.Stop(c)

	<-c
}
