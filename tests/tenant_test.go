package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

func TestTenantIsolation(t *testing.T) {
	db := orm.New(mem.New())
	modA, modB := setupTenant(t, db, "tenant-A"), setupTenant(t, db, "tenant-B")

	created, err := createWithIP(modA, devicemanager.Device{
		Name:     "Tenant A Device",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	}, "172.16.0.1")
	if err != nil {
		t.Fatalf("failed to create tenant A device: %v", err)
	}

	// 1. Tenant B no puede leer el equipo del tenant A (ni por su módulo ni pidiendo el tenant A).
	if _, err := modB.GetDevice("tenant-B", created.Id); !isErr(err, devicemanager.ErrNotFound) {
		t.Errorf("expected ErrNotFound for tenant B getting tenant A's device, got %v", err)
	}
	if _, err := modB.GetDevice("tenant-A", created.Id); !isErr(err, devicemanager.ErrNotFound) {
		t.Errorf("expected ErrNotFound when a module of tenant B is asked for tenant A, got %v", err)
	}

	// 2. Tenant B no puede actualizar, desactivar ni eliminar el equipo del tenant A.
	hijack := created
	hijack.Name = "Hijacked Name"
	if _, err := modB.UpdateDevice(hijack); err == nil {
		t.Errorf("tenant B updated tenant A's device")
	}
	if err := modB.DeactivateDevice("tenant-B", created.Id); err == nil {
		t.Errorf("tenant B deactivated tenant A's device")
	}
	if err := modB.DeleteDevice("tenant-B", created.Id); err == nil {
		t.Errorf("tenant B deleted tenant A's device")
	}
	if got, err := modA.GetDevice("tenant-A", created.Id); err != nil || got.Name != "Tenant A Device" || !got.IsActive {
		t.Errorf("tenant A's device must be untouched, got %+v err %v", got, err)
	}

	// 3. Tenant B puede registrar la MISMA ip que el tenant A: la unicidad es por tenant.
	if _, err := createWithIP(modB, devicemanager.Device{Name: "Tenant B Device", Type: devicemanager.DeviceTypeComputer, IsActive: true}, "172.16.0.1"); err != nil {
		t.Errorf("expected tenant B to reuse tenant A's ip independently, got %v", err)
	}
}
