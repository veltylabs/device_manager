package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/model"
	"webtyp.com/view"
)

// Las tres pantallas se agrupan bajo un enlace padre, validan según view.Screen y cada una tiene su propio dibujo y permiso.
func TestUI_ScreensAreDescribedWithTheirOwnIcon(t *testing.T) {
	fake := &uiFakeCaller{}
	screens := ui.Screens(view.BrowserDeps{Caller: view.NewRecords().Caller(fake)})
	if len(screens) != 1 {
		t.Fatalf("expected 1 parent screen, got %d", len(screens))
	}
	parent := screens[0]
	if err := parent.Validate(); err != nil {
		t.Fatalf("parent screen %s validation failed: %v", parent.ID, err)
	}
	if parent.ID != ui.ID || parent.Icon != laptop.Ref.ID() {
		t.Errorf("parent screen = (%q, %q), want (%q, %q)", parent.ID, parent.Icon, ui.ID, laptop.Ref.ID())
	}
	if len(parent.Children) != 3 {
		t.Fatalf("expected 3 child screens, got %d", len(parent.Children))
	}

	want := []struct {
		id         string
		icon       string
		permission string
	}{
		{ui.ID + ".devices", laptop.Ref.ID(), string(model.ResourceOf(&devicemanager.DeviceModel))},
		{ui.ZonesID, location.Ref.ID(), string(model.ResourceOf(&devicemanager.ZoneModel))},
		{ui.InterfacesID, network.Ref.ID(), string(model.ResourceOf(&devicemanager.NetworkInterfaceModel))},
	}

	for i, w := range want {
		s := parent.Children[i]
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
