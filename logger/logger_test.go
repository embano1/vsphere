package logger

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gotest.tools/v3/assert"
)

func TestGetSet(t *testing.T) {
	t.Run("returns default logger", func(t *testing.T) {
		ctx := context.Background()
		l := Get(ctx)
		assert.Assert(t, l != nil)
		assert.Equal(t, reflect.TypeOf(*l).String(), "zap.Logger")
	})

	t.Run("returns custom logger", func(t *testing.T) {
		console := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

		ws := writeSyncer{
			&bytes.Buffer{},
		}
		cfg := zapcore.NewCore(console, ws, zapcore.DebugLevel)

		l := zap.New(cfg).Named("testlogger")

		ctx := Set(context.Background(), l)
		Get(ctx).Debug("hello world")

		assert.Assert(t, strings.Contains(ws.String(), "DEBUG") == true)
		assert.Assert(t, strings.Contains(ws.String(), "testlogger") == true)
		assert.Assert(t, strings.Contains(ws.String(), "hello world") == true)
	})
}

type writeSyncer struct {
	*bytes.Buffer
}

func (t writeSyncer) Sync() error {
	return nil
}
