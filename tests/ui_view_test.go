package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/dom"
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

func TestUIView(t *testing.T) {
	fake := &uiFakeCaller{}
	ids := &mockIDGen{}

	mod, err := ui.Browser(fake, ids, "tenant-test")
	if err != nil {
		t.Fatalf("ui.Browser: %v", err)
	}

	if mod.ModelName() != ui.ID {
		t.Errorf("ModelName() = %q, want %q", mod.ModelName(), ui.ID)
	}

	if mod.Label() != ui.Label {
		t.Errorf("Label() = %q, want %q", mod.Label(), ui.Label)
	}

	v := mod.View()
	if v == nil {
		t.Fatal("View() returned nil")
	}

	if initializer, ok := v.(interface{ Init(dom.Ctx) }); ok {
		initializer.Init(nil)
	} else {
		t.Fatal("view does not implement Init method")
	}

	if len(fake.calls) != 0 {
		t.Fatal("expected no calls on Init (lazy activation)")
	}

	mod.Activate()

	if len(fake.calls) == 0 {
		t.Fatal("expected calls on Activate, got none")
	}

	lastCall := fake.calls[len(fake.calls)-1]
	expectedSuffix := ui.ID + ".list_devices"
	if len(lastCall) < len(expectedSuffix) || lastCall[len(lastCall)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("expected call ending with %q, got %q", expectedSuffix, lastCall)
	}
}

// The select options are loaded on ACTIVATION (never at creation, never at
// Init), once: a second activation does not ask for the zones again.
func TestUIView_ZoneOptionsLoadOnFirstActivation(t *testing.T) {
	fake := &uiFakeCaller{}
	mod, err := ui.Browser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatalf("ui.Browser: %v", err)
	}
	zonesOp := devicemanager.ModelName + "." + devicemanager.OpListZones
	count := func() int {
		n := 0
		for _, c := range fake.calls {
			if c == zonesOp {
				n++
			}
		}
		return n
	}
	if len(fake.calls) != 0 {
		t.Fatalf("calls before activation: %v", fake.calls)
	}
	mod.Activate()
	if count() != 1 {
		t.Fatalf("list_zones calls after first Activate = %d, want 1 (calls %v)", count(), fake.calls)
	}
	mod.Activate()
	if count() != 1 {
		t.Errorf("list_zones called again on second Activate (calls %v)", fake.calls)
	}
}

func TestUIView_ExtraScreens(t *testing.T) {
	fake := &uiFakeCaller{}
	zones, err := ui.ZonesBrowser(fake, &mockIDGen{}, "tenant-test")
	if err != nil || zones.ModelName() != ui.ZonesID {
		t.Fatalf("ZonesBrowser = %v (err %v)", zones, err)
	}
	ifaces, err := ui.InterfacesBrowser(fake, &mockIDGen{}, "tenant-test")
	if err != nil || ifaces.ModelName() != ui.InterfacesID {
		t.Fatalf("InterfacesBrowser = %v (err %v)", ifaces, err)
	}
	if len(fake.calls) != 0 {
		t.Fatalf("calls at creation: %v", fake.calls)
	}
	ifaces.Activate()
	devicesOp := devicemanager.ModelName + "." + devicemanager.OpListDevices
	found := false
	for _, c := range fake.calls {
		if c == devicesOp {
			found = true
		}
	}
	if !found {
		t.Errorf("InterfacesBrowser activation did not load the device options (calls %v)", fake.calls)
	}
}
