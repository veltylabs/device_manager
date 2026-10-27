package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

func TestIPLocator(t *testing.T) {
	db := orm.New(mem.New())
	ids := &mockIDGen{}

	dm, err := devicemanager.New(db, devicemanager.Deps{IDs: ids, TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	d1, err := createWithIP(dm, devicemanager.Device{
		TenantId: "tenant-1",
		Name:     "Box 1",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	}, "192.168.1.10")
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}

	// Crear equipo con la misma IP en otro tenant
	_, err = createWithIP(dm, devicemanager.Device{
		TenantId: "tenant-2",
		Name:     "Box 1 Tenant 2",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	}, "192.168.1.10")
	if err != nil {
		t.Fatalf("CreateDevice tenant-2: %v", err)
	}

	locator := devicemanager.IPLocator{
		Devices:  dm,
		TenantID: "tenant-1",
	}

	// IP existente en tenant-1
	id, ok := locator.FindByIP("192.168.1.10")
	if !ok || id != d1.Id {
		t.Errorf("FindByIP(192.168.1.10) = (%q, %v), want (%q, true)", id, ok, d1.Id)
	}

	// IP desconocida
	id, ok = locator.FindByIP("10.0.0.1")
	if ok || id != "" {
		t.Errorf("FindByIP(10.0.0.1) = (%q, %v), want (\"\", false)", id, ok)
	}

	// Localizador para tenant-3 (sin equipos)
	locator3 := devicemanager.IPLocator{
		Devices:  dm,
		TenantID: "tenant-3",
	}
	id, ok = locator3.FindByIP("192.168.1.10")
	if ok || id != "" {
		t.Errorf("FindByIP tenant-3 = (%q, %v), want (\"\", false)", id, ok)
	}
}

// La IP con la que se guarda un equipo y la IP que auth observa al iniciar sesión
// deben tener una única forma (input.CanonicalIP), o la misma máquina será rechazada
// dependiendo de cómo se conectó: "localhost" es ::1 desde curl, 127.0.0.1 desde Chrome.
// ClientIP de webtyp.com/auth ya devuelve la forma canónica; este módulo
// es el otro lado de esa comparación.
func TestIPLocator_CanonicalSpelling(t *testing.T) {
	db := orm.New(mem.New())
	dm, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	d, err := createWithIP(dm, devicemanager.Device{
		TenantId: "tenant-1",
		Name:     "Admin",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	}, "::1")
	if err != nil {
		t.Fatalf("createWithIP: %v", err)
	}
	ifaces, err := dm.ListNetworkInterfaces("tenant-1", d.Id)
	if err != nil || len(ifaces) != 1 || ifaces[0].Ip != "127.0.0.1" {
		t.Errorf("stored interface = %+v (err %v), want one with Ip %q", ifaces, err, "127.0.0.1")
	}

	locator := devicemanager.IPLocator{Devices: dm, TenantID: "tenant-1"}
	for _, ip := range []string{"127.0.0.1", "::1"} {
		if id, ok := locator.FindByIP(ip); !ok || id != d.Id {
			t.Errorf("FindByIP(%q) = (%q, %v), want (%q, true)", ip, id, ok, d.Id)
		}
	}

	_, err = createWithIP(dm, devicemanager.Device{
		TenantId: "tenant-1",
		Name:     "Admin again",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	}, "127.0.0.1")
	if !isErr(err, devicemanager.ErrIPAlreadyExists) {
		t.Errorf("CreateDevice same machine, other spelling: err = %v, want ErrIPAlreadyExists", err)
	}
}
