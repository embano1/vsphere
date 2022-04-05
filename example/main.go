package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/embano1/vsphere/client"
	"github.com/embano1/vsphere/event"
	"github.com/embano1/vsphere/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	l, err := cfg.Build()
	if err != nil {
		panic("create logger: " + err.Error())
	}
	l = l.Named("demo-app")

	// inject and overwrite default logger
	ctx = logger.Set(ctx, l)

	l.Debug("creating vsphere client")
	c, err := client.New(ctx)
	if err != nil {
		l.Fatal("could not create vsphere client", zap.Error(err))
	}
	defer c.Logout()

	// show how to use filters
	filterEvents := []string{"UserLoginSessionEvent", "VmPoweredOnEvent", "DrsVmPoweredOnEvent"}
	filters := []event.Filter{event.WithEventTypeID(filterEvents)}

	// retrieve (filtered) events for all vCenter objects
	root := c.SOAP.ServiceContent.RootFolder
	collector, err := event.NewHistoryCollector(ctx, c.Events, root, filters...)
	if err != nil {
		l.Fatal("could not create event stream", zap.Error(err))
	}

	const (
		batch        = 10
		pollInterval = time.Second * 3
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	l.Info("starting event stream", zap.Any("forEvents", filterEvents))
LOOP:
	for {
		select {
		case <-ctx.Done():
			l.Debug("shutting down")
			break LOOP

		case <-ticker.C:
			l.Debug("retrieving events")
			events, err := collector.ReadNextEvents(ctx, batch)
			if err != nil && !errors.Is(err, context.Canceled) {
				l.Fatal("could not read events", zap.Error(err))
			}

			if len(events) == 0 {
				l.Debug("no new events")
				continue
			}

			for _, e := range events {
				l.Info("retrieved new event", zap.Any("event", e))
			}
		}
	}

	l.Debug("shutdown complete")
}
