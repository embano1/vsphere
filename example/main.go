package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"

	"github.com/embano1/vsphere/client"
	"github.com/embano1/vsphere/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	l, err := zap.NewDevelopment()
	if err != nil {
		panic("create logger: " + err.Error())
	}

	// inject and overwrite default logger
	ctx = logger.Set(ctx, l)

	l.Debug("creating vsphere client")
	c, err := client.New(ctx)
	if err != nil {
		l.Fatal("create vsphere client", zap.Error(err))
	}

	mgr := event.NewManager(c.SOAP.Client)
	objs := []types.ManagedObjectReference{c.SOAP.ServiceContent.RootFolder}
	handler := func(reference types.ManagedObjectReference, events []types.BaseEvent) error {
		for _, e := range events {
			l.Debug("new event", zap.Any("event", e))
		}
		return nil
	}

	l.Debug("starting event stream")
	if err = mgr.Events(ctx, objs, 10, true, false, handler); err != nil {
		l.Fatal("stream events", zap.Error(err))
	}
	l.Debug("shutdown complete")
}
