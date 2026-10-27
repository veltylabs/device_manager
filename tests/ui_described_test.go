package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/view"
)

// Las tres pantallas se explican solas, validan según view.Screen y cada una tiene su propio dibujo y permiso.
func TestUI_ScreensAreDescribedWithTheirOwnIcon(t *testing.T) {
	fake := &uiFakeCaller{}
	screens := ui.Screens(view.NewRecords().Caller(fake))
	if len(screens) != 3 {
		t.Fatalf("expected 3 screens, got %d", len(screens))
	}

	want := []struct {
		id         string
		icon       string
		permission string
	}{
		{ui.ID, laptop.Ref.ID(), string(devicemanager.ResourceDevice)},
		{ui.ZonesID, location.Ref.ID(), string(devicemanager.ResourceZone)},
		{ui.InterfacesID, network.Ref.ID(), string(devicemanager.ResourceNetworkInterface)},
	}

	for i, w := range want {
		s := screens[i]
		if err := s.Validate(); err != nil {
			t.Fatalf("screen %s validation failed: %v", s.ID, err)
		}
		if s.ID != w.id {
			t.Errorf("screen %d ID = %q, want %q", i, s.ID, w.id)
		}
		if s.Icon != w.icon {
			t.Errorf("screen %d Icon = %q, want %q", i, s.Icon, w.icon)
		}
		if string(s.Permission) != w.permission {
			t.Errorf("screen %d Permission = %q, want %q", i, s.Permission, w.permission)
		}
		if s.Presenter.Description() == "" {
			t.Errorf("screen %d presenter must have a non-empty text", i)
		}
	}
}
