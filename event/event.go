package event

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/types"
)

// NewHistoryCollector creates a new event collector for the specified entity.
// By default, events for the entity and all (indirect) children (if any) are
// retrieved and event collection starts at "now".
func NewHistoryCollector(ctx context.Context, mgr *event.Manager, entity types.ManagedObjectReference, filters ...Filter) (*event.HistoryCollector, error) {
	f := defaultFilters
	f = append(f, filters...)
	spec, err := createSpec(entity, f)
	if err != nil {
		return nil, fmt.Errorf("create filter spec: %w", err)
	}
	return mgr.CreateCollectorForEvents(ctx, *spec)
}

func createSpec(entity types.ManagedObjectReference, filters []Filter) (*types.EventFilterSpec, error) {
	spec := types.EventFilterSpec{
		Entity: &types.EventFilterSpecByEntity{
			Entity: entity,
		},
	}

	for _, f := range filters {
		if err := f(&spec); err != nil {
			return nil, fmt.Errorf("filter spec invalid: %w", err)
		}
	}

	return &spec, nil
}

// Details contains the type and class of an event received from vCenter
// supported event classes: event, eventex, extendedevent.
//
// Class to type mapping:
// event: retrieved from event Class, e.g.VmPoweredOnEvent
// eventex: retrieved from EventTypeId
// extendedevent: retrieved from EventTypeId
type Details struct {
	Class string
	Type  string
}

// GetDetails retrieves the underlying vSphere event class and name for the
// given BaseEvent, e.g. VmPoweredOnEvent (event) or
// com.vmware.applmgmt.backup.job.failed.event (extendedevent)
func GetDetails(event types.BaseEvent) Details {
	var details Details

	switch e := event.(type) {
	case *types.EventEx:
		details.Class = "eventex"
		details.Type = e.EventTypeId
	case *types.ExtendedEvent:
		details.Class = "extendedevent"
		details.Type = e.EventTypeId
	default:
		t := reflect.TypeOf(event).Elem().Name()
		details.Class = "event"
		details.Type = t
	}

	return details
}
