package event

import (
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	// signal unstable event API for converting vSphere events to CE
	ceEventTypeFormat = "com.vmware.vsphere.%s.v0"
)

// ToCloudEvent transforms the vSphere event into a CloudEvent. The full vSphere
// event is JSON-encoded and available in the data field. Extensions sets the
// map keys and values as CloudEvent extensions. Optional, i.e. can be nil. If
// specified, extensions must contain valid CloudEvent extensions.
func ToCloudEvent(source string, be types.BaseEvent, extensions map[string]string) (ce.Event, error) {
	details := GetDetails(be)

	e := ce.NewEvent()
	e.SetSource(source)
	e.SetID(fmt.Sprintf("%d", be.GetEvent().Key))
	e.SetType(fmt.Sprintf(ceEventTypeFormat, details.Type))
	e.SetTime(be.GetEvent().CreatedTime)

	for k, v := range extensions {
		e.SetExtension(k, v)
	}

	if err := e.SetData(ce.ApplicationJSON, be); err != nil {
		return ce.Event{}, fmt.Errorf("marshal vsphere event to cloudevent data: %w", err)
	}

	if err := e.Validate(); err != nil {
		return ce.Event{}, fmt.Errorf("convert vsphere event to cloudevent: %w", err)
	}

	return e, nil
}
