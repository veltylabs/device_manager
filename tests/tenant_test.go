package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

func TestTenantIsolation(t *testing.T) {
	db := orm.New(mem.New())
	m, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-A"})
	if err != nil {
		t.Fatal(err)
	}

	tenantA := "tenant-A"
	tenantB := "tenant-B"

	created, err := createWithIP(m, devicemanager.Device{
		TenantId: tenantA,
		Name:     "Tenant A Device",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	}, "172.16.0.1")
	if err != nil {
		t.Fatalf("failed to create tenant A device: %v", err)
	}

	// 1. Tenant B NO debe poder hacer Get al equipo del tenant A
	if _, err := m.GetDevice(tenantB, created.Id); !isErr(err, devicemanager.ErrNotFound) {
		t.Errorf("expected ErrNotFound for tenant B getting tenant A's device, got %v", err)
	}

	// 2. Tenant B NO debe poder hacer Update al equipo del tenant A
	hijack := created
	hijack.TenantId = tenantB
	hijack.Name = "Hijacked Name"
	if _, err := m.UpdateDevice(hijack); !isErr(err, devicemanager.ErrNotFound) {
		t.Errorf("expected ErrNotFound for tenant B updating tenant A's device, got %v", err)
	}

	// 3. Tenant B NO debe poder desactivar (Deactivate) el equipo del tenant A
	if err := m.DeactivateDevice(tenantB, created.Id); !isErr(err, devicemanager.ErrNotFound) {
		t.Errorf("expected ErrNotFound for tenant B deactivating tenant A's device, got %v", err)
	}

	// 4. Tenant B NO debe poder eliminar (Delete) el equipo del tenant A
	if err := m.DeleteDevice(tenantB, created.Id); !isErr(err, devicemanager.ErrNotFound) {
		t.Errorf("expected ErrNotFound for tenant B deleting tenant A's device, got %v", err)
	}

	// 5. Que Tenant B registre la MISMA ip que el tenant A debe estar permitido (la unicidad es por tenant)
	if _, err := createWithIP(m, devicemanager.Device{TenantId: tenantB, Name: "Tenant B Device", Type: devicemanager.DeviceTypeComputer, IsActive: true}, "172.16.0.1"); err != nil {
		t.Errorf("expected tenant B to reuse tenant A's ip independently, got %v", err)
	}
}
