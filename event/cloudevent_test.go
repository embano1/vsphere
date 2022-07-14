package event

import (
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware/govmomi/vim25/types"
	"gotest.tools/v3/assert"
)

func Test_ToCloudEvent(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		extensions map[string]string
		want       func(time time.Time) ce.Event
		wantErr    string
	}{
		{
			name:       "fails to create event with invalid extension",
			source:     "/testsource",
			extensions: map[string]string{"123-hello": "invalid"},
			want:       nil,
			wantErr:    "bad key",
		},
		{
			name:    "fails to create event with invalid source",
			source:  "",
			want:    nil,
			wantErr: "source: REQUIRED",
		},
		{
			name:   "creates valid cloud event",
			source: "/testsource",
			want: func(time time.Time) ce.Event {
				e := ce.NewEvent()
				e.SetID("1234")
				e.SetSource("/testsource")
				e.SetType("com.vmware.vsphere.VmPoweredOnEvent.v0")
				e.SetTime(time)

				err := e.SetData(ce.ApplicationJSON, newVSphereEvent(time))
				assert.NilError(t, err)
				return e
			},
			wantErr: "",
		},
		{
			name:       "creates valid cloud event with extensions",
			source:     "/testsource",
			extensions: map[string]string{"vsphereclass": "event"},
			want: func(time time.Time) ce.Event {
				e := ce.NewEvent()
				e.SetID("1234")
				e.SetSource("/testsource")
				e.SetType("com.vmware.vsphere.VmPoweredOnEvent.v0")
				e.SetTime(time)
				e.SetExtension("vsphereclass", "event")

				err := e.SetData(ce.ApplicationJSON, newVSphereEvent(time))
				assert.NilError(t, err)
				return e
			},
			wantErr: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()

			got, err := ToCloudEvent(tc.source, newVSphereEvent(now), tc.extensions)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tc.want(now), got)
			}
		})
	}
}

func newVSphereEvent(time time.Time) types.BaseEvent {
	return &types.VmPoweredOnEvent{
		VmEvent: types.VmEvent{
			Event: types.Event{
				Key:             1234,
				ChainId:         1234,
				CreatedTime:     time,
				UserName:        "test-user",
				Datacenter:      nil,
				ComputeResource: nil,
				Host:            nil,
				Vm: &types.VmEventArgument{
					Vm: types.ManagedObjectReference{
						Type:  "VirtualMachine",
						Value: "vm-1234",
					},
				},
				FullFormattedMessage: "some test message",
			},
			Template: false,
		},
	}
}
