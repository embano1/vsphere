package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"go.uber.org/zap"

	"github.com/embano1/vsphere/client"
	"github.com/embano1/vsphere/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l := logger.Get(ctx)

	if info, ok := debug.ReadBuildInfo(); ok {
		l.Info("go runtime", zap.String("info", info.String()))
	}
	l.Info("go environment", zap.Any("env", os.Environ()))

	c, err := client.New(ctx)
	if err != nil {
		l.Fatal("create client", zap.Error(err))
	}
	defer func() {
		_ = c.Logout()
	}()

	v := c.SOAP.Version
	if v == "" {
		l.Fatal("invalid version received", zap.String("version", v))
	}
	l.Info("successfully connected to vsphere", zap.String("version", v))
}
