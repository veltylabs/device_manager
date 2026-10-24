//go:build browser

package tests

import (
	"testing"

	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/layout/platformd"
)

// Las tres pantallas se explican solas y cada una tiene su propio dibujo.
func TestUI_ScreensAreDescribedWithTheirOwnIcon(t *testing.T) {
	fake := &uiFakeCaller{}
	devices, err := ui.Browser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatal(err)
	}
	zones, err := ui.ZonesBrowser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatal(err)
	}
	ifaces, err := ui.InterfacesBrowser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		screen platformd.UIModule
		icon   string
	}{{devices, "laptop"}, {zones, "location"}, {ifaces, "network"}}
	for _, w := range want {
		d, ok := w.screen.(platformd.Described)
		if !ok || d.Description() == "" {
			t.Errorf("%s must implement platformd.Described with a non-empty text", w.screen.ModelName())
		}
		if got := w.screen.Icon().ID(); got != w.icon {
			t.Errorf("%s icon = %q, want %q", w.screen.ModelName(), got, w.icon)
		}
	}
}
