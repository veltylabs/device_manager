package tests

import (
	"testing"

	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/model"
)

type uiFakeCaller struct {
	calls []string
}

func (f *uiFakeCaller) Call(op string, args model.Encodable, out model.Decodable, done func(error)) {
	f.calls = append(f.calls, op)
	if done != nil {
		done(nil)
	}
}

func (f *uiFakeCaller) Dispatch(op string, args model.Encodable) {}

func TestUIView_Lookups(t *testing.T) {
	fake := &uiFakeCaller{}
	screens := ui.Screens(fake)

	// Dispositivos debe tener lookup en zone_id
	devScreen := screens[0]
	if len(devScreen.Lookups) != 1 {
		t.Fatalf("expected 1 lookup on devices screen, got %d", len(devScreen.Lookups))
	}
	if devScreen.Lookups[0].Field != "zone_id" {
		t.Errorf("lookup field = %q, want zone_id", devScreen.Lookups[0].Field)
	}
	if devScreen.Lookups[0].Source == nil {
		t.Errorf("lookup source must not be nil")
	}

	// Interfaces debe tener lookup en device_id
	ifaceScreen := screens[2]
	if len(ifaceScreen.Lookups) != 1 {
		t.Fatalf("expected 1 lookup on interfaces screen, got %d", len(ifaceScreen.Lookups))
	}
	if ifaceScreen.Lookups[0].Field != "device_id" {
		t.Errorf("lookup field = %q, want device_id", ifaceScreen.Lookups[0].Field)
	}
	if ifaceScreen.Lookups[0].Source == nil {
		t.Errorf("lookup source must not be nil")
	}
}
