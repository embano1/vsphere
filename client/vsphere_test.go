package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/vmware/govmomi/simulator"
	_ "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vim25"
	"gotest.tools/v3/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("fails to create client", func(t *testing.T) {
		testCases := []struct {
			name    string
			env     map[string]string
			wantErr string
		}{
			{
				name: "credentials not found",
				env: map[string]string{
					// if empty, use val injected from test
					"VCENTER_URL":         "",
					"VCENTER_INSECURE":    "true",
					"VCENTER_SECRET_PATH": "/var/bindings/vsphere",
				},
				wantErr: "no such file",
			},
			{
				name: "certificate error",
				env: map[string]string{
					// if empty, use val injected from test
					"VCENTER_URL":         "",
					"VCENTER_INSECURE":    "",
					"VCENTER_SECRET_PATH": "",
				},
				wantErr: "certificate signed by unknown authority",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				dir := tempDir(t)

				t.Cleanup(func() {
					err := os.RemoveAll(dir)
					assert.NilError(t, err)
				})

				simulator.Run(func(ctx context.Context, vimclient *vim25.Client) error {
					// inject test dir
					t.Setenv("VCENTER_SECRET_PATH", dir)
					t.Setenv("VCENTER_URL", vimclient.URL().String())
					t.Setenv("VCENTER_INSECURE", "false")

					for env, v := range tc.env {
						if v != "" {
							t.Setenv(env, v)
						}
					}

					c, err := New(ctx)
					assert.ErrorContains(t, err, tc.wantErr)
					assert.DeepEqual(t, c, (*Client)(nil))

					return nil
				})
			})
		}
	})

	t.Run("successfully creates client", func(t *testing.T) {
		dir := tempDir(t)

		t.Cleanup(func() {
			err := os.RemoveAll(dir)
			assert.NilError(t, err)
		})

		simulator.Run(func(ctx context.Context, vimclient *vim25.Client) error {
			t.Setenv("VCENTER_URL", vimclient.URL().String())
			t.Setenv("VCENTER_INSECURE", "true")
			t.Setenv("VCENTER_SECRET_PATH", dir)

			c, err := New(ctx)
			assert.NilError(t, err)
			assert.Assert(t, c.SOAP != nil)
			assert.Assert(t, c.REST != nil)
			assert.Assert(t, c.Tags != nil)
			assert.Assert(t, c.Tasks != nil)
			assert.Assert(t, c.Events != nil)

			err = c.Logout()
			assert.NilError(t, err)

			return nil
		})
	})
}

func tempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "")
	assert.NilError(t, err)

	f, err := os.Create(filepath.Join(dir, userFileKey))
	assert.NilError(t, err)

	_, err = f.Write([]byte("user"))
	assert.NilError(t, err)
	assert.NilError(t, f.Close())

	f, err = os.Create(filepath.Join(dir, passwordFileKey))
	assert.NilError(t, err)

	_, err = f.Write([]byte("pass"))
	assert.NilError(t, err)
	assert.NilError(t, f.Close())

	return dir
}
