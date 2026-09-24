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
