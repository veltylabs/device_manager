package tests

import (
	"testing"

	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage/mem"
	devicemanager "github.com/veltylabs/device_manager"
)

func TestTenantIsolation(t *testing.T) {
	db := orm.New(mem.New())
	m, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}})
	if err != nil {
		t.Fatal(err)
	}

	tenantA := "tenant-A"
	tenantB := "tenant-B"

	created, err := m.CreateDevice(devicemanager.Device{
		TenantId: tenantA,
		Name:     "Tenant A Device",
		Ip:       "172.16.0.1",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("failed to create tenant A device: %v", err)
	}

	// 1. Tenant B must NOT be able to Get tenant A's device
	if _, err := m.GetDevice(tenantB, created.Id); err != devicemanager.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B getting tenant A's device, got %v", err)
	}

	// 2. Tenant B must NOT be able to Update tenant A's device
	hijack := created
	hijack.TenantId = tenantB
	hijack.Name = "Hijacked Name"
	if _, err := m.UpdateDevice(hijack); err != devicemanager.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B updating tenant A's device, got %v", err)
	}

	// 3. Tenant B must NOT be able to Deactivate tenant A's device
	if err := m.DeactivateDevice(tenantB, created.Id); err != devicemanager.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B deactivating tenant A's device, got %v", err)
	}

	// 4. Tenant B must NOT be able to Delete tenant A's device
	if err := m.DeleteDevice(tenantB, created.Id); err != devicemanager.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B deleting tenant A's device, got %v", err)
	}

	// 5. Tenant B registering the SAME ip as tenant A must be allowed (uniqueness is per tenant)
	if _, err := m.CreateDevice(devicemanager.Device{TenantId: tenantB, Name: "Tenant B Device", Ip: "172.16.0.1", Type: devicemanager.DeviceTypeComputer, IsActive: true}); err != nil {
		t.Errorf("expected tenant B to reuse tenant A's ip independently, got %v", err)
	}
}
