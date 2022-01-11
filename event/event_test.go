package event

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"gotest.tools/v3/assert"
)

func Test_NewHistoryCollector(t *testing.T) {
	simulator.Run(func(ctx context.Context, client *vim25.Client) error {
		f := WithTime(&types.EventFilterSpecByTime{
			BeginTime: types.NewTime(time.Now().UTC().Add(-5 * time.Minute)), // since start
		})

		mgr := event.NewManager(client)
		collector, err := NewHistoryCollector(ctx, mgr, client.ServiceContent.RootFolder, f)
		assert.NilError(t, err)

		events, err := collector.ReadNextEvents(ctx, 100)
		assert.NilError(t, err)
		assert.Assert(t, len(events) > 0)

		return nil
	})
}

func Test_createSpec(t *testing.T) {
	const (
		notNilErr = "must not be nil"
	)

	entity := types.ManagedObjectReference{
		Type:  "host-1",
		Value: "Host",
	}

	t.Run("fails when fs input is invalid", func(t *testing.T) {
		testCases := []struct {
			name    string
			f       Filter
			wantErr string
		}{
			{
				name:    "EventTypeID is nil",
				f:       WithEventTypeID(nil),
				wantErr: notNilErr,
			},
			{
				name:    "Time is nil",
				f:       WithTime(nil),
				wantErr: notNilErr,
			},
			{
				name:    "Username is nil",
				f:       WithUsername(nil),
				wantErr: notNilErr,
			},
			{
				name:    "MaxCount is 0",
				f:       WithMaxCount(0),
				wantErr: "be greater than 0",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := createSpec(entity, []Filter{tc.f})
				assert.ErrorContains(t, err, tc.wantErr)
			})
		}
	})

	t.Run("creates spec with defaults", func(t *testing.T) {
		spec, err := createSpec(entity, defaultFilters)
		assert.NilError(t, err)
		assert.Assert(t, spec.Entity != nil)
		assert.Assert(t, spec.Time != nil)
		assert.Assert(t, spec.UserName == nil)
		assert.Assert(t, spec.EventTypeId == nil)
		assert.Assert(t, spec.MaxCount == 0)
	})

	t.Run("creates custom spec", func(t *testing.T) {
		now := time.Now().UTC()
		testCases := []struct {
			name string
			fs   []Filter
			want *types.EventFilterSpec
		}{
			{
				name: "begins -10m ago",
				fs: []Filter{
					WithTime(&types.EventFilterSpecByTime{
						BeginTime: types.NewTime(now),
						EndTime:   nil,
					}),
				},
				want: &types.EventFilterSpec{
					Entity: &types.EventFilterSpecByEntity{
						Entity:    entity,
						Recursion: types.EventFilterSpecRecursionOptionAll,
					},
					Time: &types.EventFilterSpecByTime{
						BeginTime: types.NewTime(now),
					},
				},
			}, {
				name: "begins -1h ago, no recursion",
				fs: []Filter{
					WithTime(&types.EventFilterSpecByTime{
						BeginTime: types.NewTime(now.Add(-1 * time.Hour)),
						EndTime:   nil,
					}),
					WithRecursion(types.EventFilterSpecRecursionOptionSelf),
				},
				want: &types.EventFilterSpec{
					Entity: &types.EventFilterSpecByEntity{
						Entity:    entity,
						Recursion: types.EventFilterSpecRecursionOptionSelf,
					},
					Time: &types.EventFilterSpecByTime{
						BeginTime: types.NewTime(now.Add(-1 * time.Hour)),
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fs := defaultFilters
				fs = append(fs, tc.fs...)
				spec, err := createSpec(entity, fs)
				assert.NilError(t, err)
				assert.DeepEqual(t, spec, tc.want)
			})
		}
	})
}

func Test_GetDetails(t *testing.T) {
	testCases := []struct {
		name  string
		event types.BaseEvent
		want  Details
	}{
		{
			name:  "VmPoweredOnEvent",
			event: &types.VmPoweredOnEvent{},
			want: Details{
				Class: "event",
				Type:  "VmPoweredOnEvent",
			},
		},
		{
			name: "Extended Event",
			event: &types.ExtendedEvent{
				EventTypeId: "com.backup.job.succeeded",
			},
			want: Details{
				Class: "extendedevent",
				Type:  "com.backup.job.succeeded",
			},
		},
		{
			name: "EventEx",
			event: &types.EventEx{
				EventTypeId: "com.appliance.shutdown.succeeded",
			},
			want: Details{
				Class: "eventex",
				Type:  "com.appliance.shutdown.succeeded",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := GetDetails(tc.event)
			assert.DeepEqual(t, d, tc.want)
		})
	}
}
