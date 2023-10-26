package client_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"go.uber.org/zap"

	"github.com/embano1/vsphere/logger"

	"github.com/embano1/vsphere/client"
)

func ExampleNew() {
	// use vcenter simulator
	simulator.Run(func(ctx context.Context, simClient *vim25.Client) error {
		l, err := setup(simClient.URL().String())
		if err != nil {
			return fmt.Errorf("setup environment: %w", err)
		}

		ctx = logger.Set(ctx, l)
		c, err := client.New(ctx)
		if err != nil {
			l.Fatal("create vsphere client", zap.Error(err))
		}

		defer func() {
			if err = c.Logout(); err != nil {
				l.Warn("logout", zap.Error(err))
			}
		}()

		l.Info("connected to vcenter", zap.String("version", c.SOAP.Version))
		return nil
	})

	// 	Output: INFO connected to vcenter {"version": "8.0.2.0"}
}

// this is only needed for the example. In a real deployment, e.g. Kubernetes
// the secret and environment variables would be injected.
func setup(url string) (*zap.Logger, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}

	f, err := os.Create(filepath.Join(dir, "username"))
	if err != nil {
		return nil, fmt.Errorf("create user file: %w", err)
	}

	_, err = f.Write([]byte("usr"))
	if err != nil {
		return nil, fmt.Errorf("write to user file: %w", err)
	}
	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("close user file: %w", err)
	}

	f, err = os.Create(filepath.Join(dir, "password"))
	if err != nil {
		return nil, fmt.Errorf("create password file: %w", err)
	}

	_, err = f.Write([]byte("pass"))
	if err != nil {
		return nil, fmt.Errorf("write to password file: %w", err)
	}

	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("close password file: %w", err)
	}

	env := map[string]string{
		"VCENTER_URL":         url,
		"VCENTER_INSECURE":    "true",
		"VCENTER_SECRET_PATH": dir,
	}

	for e, v := range env {
		if err = os.Setenv(e, v); err != nil {
			return nil, fmt.Errorf("set %q env var: %w", e, err)
		}
	}

	// create logger
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.TimeKey = ""   // no timestamps
	cfg.EncoderConfig.CallerKey = "" // no func lines
	cfg.EncoderConfig.ConsoleSeparator = " "

	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	return l, nil
}
