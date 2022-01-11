package event

import (
	"fmt"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

// Filter is a filter applied to the event filter spec. See vSphere API
// documentation for Details on the specific fields.
type Filter func(f *types.EventFilterSpec) error

// WithEventTypeID limits the set of collected events to those specified types
func WithEventTypeID(ids []string) Filter {
	return func(f *types.EventFilterSpec) error {
		if ids == nil {
			return fmt.Errorf("types filter must not be nil")
		}
		f.EventTypeId = ids
		return nil
	}
}

// WithMaxCount specifies the maximum number of returned events
func WithMaxCount(count uint32) Filter {
	return func(f *types.EventFilterSpec) error {
		if count == 0 {
			return fmt.Errorf("count must be greater than 0")
		}
		f.MaxCount = int32(count)
		return nil
	}
}

// WithRecursion specifies whether events should be received only for the
// specified object, including its direct children or all children
func WithRecursion(r types.EventFilterSpecRecursionOption) Filter {
	return func(f *types.EventFilterSpec) error {
		f.Entity.Recursion = r
		return nil
	}
}

// WithTime filters events based on time
func WithTime(time *types.EventFilterSpecByTime) Filter {
	return func(f *types.EventFilterSpec) error {
		if time == nil {
			return fmt.Errorf("time filter must not be nil")
		}
		f.Time = time
		return nil
	}
}

// WithUsername filters events based on username
func WithUsername(u *types.EventFilterSpecByUsername) Filter {
	return func(f *types.EventFilterSpec) error {
		if u == nil {
			return fmt.Errorf("username filter must not be nil")
		}
		f.UserName = u
		return nil
	}
}

var defaultFilters = []Filter{
	WithRecursion(types.EventFilterSpecRecursionOptionAll),
	// explicitly start at "now"
	WithTime(&types.EventFilterSpecByTime{BeginTime: types.NewTime(time.Now().UTC())}),
}
