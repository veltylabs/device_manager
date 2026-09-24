package tests

import (
	"testing"

	"github.com/veltylabs/device_manager/seed"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	devicemanager "github.com/veltylabs/device_manager"
)

func TestSeedLoad(t *testing.T) {
	db := orm.New(mem.New())
	ids := &mockIDGen{}

	dm, err := devicemanager.New(db, devicemanager.Deps{IDs: ids, TenantID: "tenant-demo"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	data, err := seed.Load(dm, "tenant-demo")
	if err != nil {
		t.Fatalf("seed.Load: %v", err)
	}

	if len(data.Devices) != 3 {
		t.Errorf("got %d devices, want 3", len(data.Devices))
	}

	devices, err := dm.ListDevices("tenant-demo", devicemanager.DeviceFilter{})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(devices) != 3 {
		t.Errorf("ListDevices got %d devices, want 3", len(devices))
	}
}
