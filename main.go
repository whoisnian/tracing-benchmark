package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whoisnian/tracing-benchmark/global"
	"github.com/whoisnian/tracing-benchmark/router"
)

func main() {
	global.SetupConfig()
	global.SetupLogger()
	global.LOG.Info(fmt.Sprintf("setup config successfully: %+v", global.CFG))

	if global.CFG.Version {
		fmt.Printf("%s %s(%s)\n", global.AppName, global.Version, global.BuildTime)
		return
	}

	MustMatchTracer()
	global.SetupTracer()
	global.LOG.Info("setup tracer successfully")
	global.SetupMetrics()
	global.LOG.Info("setup metrics successfully")

	global.SetupRedis()
	global.LOG.Info("setup redis successfully")
	global.SetupMySQL()
	global.LOG.Info("setup mysql successfully")

	server := &http.Server{
		Addr:              global.CFG.ListenAddr,
		Handler:           router.Setup(),
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 180,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}
	go func() {
		global.LOG.Info("service is starting: " + global.CFG.ListenAddr)
		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			global.LOG.Warn("service is shutting down")
		} else {
			global.LOG.Error(err.Error())
			os.Exit(1)
		}
	}()

	waitFor(syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		global.LOG.Warn(err.Error())
	}
	if err := global.TR.Shutdown(ctx); err != nil {
		global.LOG.Warn(err.Error())
	}
	global.LOG.Info("service has been shut down")
}

func waitFor(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	defer signal.Stop(c)

	<-c
}
