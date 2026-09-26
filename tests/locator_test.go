package tests

import (
	"testing"

	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	devicemanager "github.com/veltylabs/device_manager"
)

func TestIPLocator(t *testing.T) {
	db := orm.New(mem.New())
	ids := &mockIDGen{}

	dm, err := devicemanager.New(db, devicemanager.Deps{IDs: ids, TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	d1, err := dm.CreateDevice(devicemanager.Device{
		TenantId: "tenant-1",
		Name:     "Box 1",
		Ip:       "192.168.1.10",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}

	// Create device with same IP on another tenant
	_, err = dm.CreateDevice(devicemanager.Device{
		TenantId: "tenant-2",
		Name:     "Box 1 Tenant 2",
		Ip:       "192.168.1.10",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateDevice tenant-2: %v", err)
	}

	locator := devicemanager.IPLocator{
		Devices:  dm,
		TenantID: "tenant-1",
	}

	// Existing IP on tenant-1
	id, ok := locator.FindByIP("192.168.1.10")
	if !ok || id != d1.Id {
		t.Errorf("FindByIP(192.168.1.10) = (%q, %v), want (%q, true)", id, ok, d1.Id)
	}

	// Unknown IP
	id, ok = locator.FindByIP("10.0.0.1")
	if ok || id != "" {
		t.Errorf("FindByIP(10.0.0.1) = (%q, %v), want (\"\", false)", id, ok)
	}

	// Locator for tenant-3 (no devices)
	locator3 := devicemanager.IPLocator{
		Devices:  dm,
		TenantID: "tenant-3",
	}
	id, ok = locator3.FindByIP("192.168.1.10")
	if ok || id != "" {
		t.Errorf("FindByIP tenant-3 = (%q, %v), want (\"\", false)", id, ok)
	}
}

// The IP a device is stored under and the IP auth observes at login must be
// one spelling (input.CanonicalIP), or the same machine is rejected depending
// on how it connected: "localhost" is ::1 from curl, 127.0.0.1 from Chrome.
// webtyp.com/auth's ClientIP already returns the canonical form; this module
// is the other side of that comparison.
func TestIPLocator_CanonicalSpelling(t *testing.T) {
	db := orm.New(mem.New())
	dm, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	d, err := dm.CreateDevice(devicemanager.Device{
		TenantId: "tenant-1",
		Name:     "Admin",
		Ip:       "::1",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if d.Ip != "127.0.0.1" {
		t.Errorf("stored Ip = %q, want %q", d.Ip, "127.0.0.1")
	}

	locator := devicemanager.IPLocator{Devices: dm, TenantID: "tenant-1"}
	for _, ip := range []string{"127.0.0.1", "::1"} {
		if id, ok := locator.FindByIP(ip); !ok || id != d.Id {
			t.Errorf("FindByIP(%q) = (%q, %v), want (%q, true)", ip, id, ok, d.Id)
		}
	}

	_, err = dm.CreateDevice(devicemanager.Device{
		TenantId: "tenant-1",
		Name:     "Admin again",
		Ip:       "127.0.0.1",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != devicemanager.ErrIPAlreadyExists {
		t.Errorf("CreateDevice same machine, other spelling: err = %v, want ErrIPAlreadyExists", err)
	}
}
