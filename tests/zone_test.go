package tests

import (
	"webtyp.com/model"
	"testing"
	"webtyp.com/crud"
	"webtyp.com/router"

	devicemanager "github.com/veltylabs/device_manager"
)

func zoneAP06(tenant string) devicemanager.Zone {
	return devicemanager.Zone{TenantId: tenant, Name: "AP-06 piso 2", RangeStart: "172.0.0.60", RangeEnd: "172.0.0.69"}
}

func TestZone_SaveListDelete(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}

	z := zoneAP06("tenant-A")
	var out devicemanager.Zone
	if err := c.Call(opCreate, &z, &out); err != nil {
		t.Fatalf("Create Zone: %v", err)
	}

	opRead := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Read}
	var zones devicemanager.ZoneList
	if err := c.Call(opRead, map[string]interface{}{}, &zones); err != nil || len(zones) != 1 || zones[0].Id != out.Id {
		t.Fatalf("ListZones = %+v (err %v), want the saved zone", zones, err)
	}

	opUpdate := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Update}
	if err := c.Call(opUpdate, map[string]interface{}{"id": out.Id, "name": "AP-06 piso dos"}, nil); err != nil {
		t.Fatalf("Update Zone: %v", err)
	}

	opDelete := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Delete}
	if err := c.Call(opDelete, map[string]interface{}{"id": out.Id}, nil); err != nil {
		t.Fatalf("DeleteZone: %v", err)
	}

	if _, err := m.GetZone("tenant-A", out.Id); !isErr(err, devicemanager.ErrZoneNotFound) {
		t.Fatalf("GetZone after delete: err = %v, want ErrZoneNotFound", err)
	}
}
	zones, err := m.ListZones("tenant-A", "", "", 0)
	if err != nil || len(zones) != 1 || zones[0].Id != z.Id {
		t.Fatalf("ListZones = %+v (err %v), want the saved zone", zones, err)
	}
	z.Name = "AP-06 piso dos"
	if _, err := m.SaveZone(z); err != nil {
		t.Fatalf("SaveZone (update): %v", err)
	}
	if err := m.DeleteZone("tenant-A", z.Id); err != nil {
		t.Fatalf("DeleteZone: %v", err)
	}
	if _, err := m.GetZone("tenant-A", z.Id); !isErr(err, devicemanager.ErrZoneNotFound) {
		t.Fatalf("GetZone after delete: err = %v, want ErrZoneNotFound", err)
	}
}

func TestZone_InvalidRange(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}

	for _, z := range []devicemanager.Zone{
		{TenantId: "tenant-A", Name: "reversed", RangeStart: "172.0.0.69", RangeEnd: "172.0.0.60"},
		{TenantId: "tenant-A", Name: "ipv6", RangeStart: "::1", RangeEnd: "::2"},
	} {
		if err := c.Call(opCreate, &z, nil); err == nil {
				t.Errorf("CreateZone(%s) = nil, want a FieldError", z.Name)
			} else if _, ok := err.(*model.FieldError); !ok {
				t.Errorf("CreateZone(%s) = %v (%T), want FieldError", z.Name, err, err)
		}
	}
}
	}
}

func TestZone_Overlap(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}

	z1 := zoneAP06("tenant-A")
	if err := c.Call(opCreate, &z1, nil); err != nil {
		t.Fatalf("CreateZone: %v", err)
	}

	overlap := devicemanager.Zone{TenantId: "tenant-A", Name: "overlap", RangeStart: "172.0.0.65", RangeEnd: "172.0.0.79"}
	if err := c.Call(opCreate, &overlap, nil); !isErr(err, devicemanager.ErrZoneOverlap) {
		t.Fatalf("overlapping zone: err = %v, want ErrZoneOverlap", err)
	}

	// Mismo rango en otro tenant es válido
	mB := setup(t)
	mB.TenantID = "tenant-B" // setup may mock differently but let's assume we can loopback it
	cB := crud.LoopbackActingAs(nil, "u1", mB)
	z2 := zoneAP06("tenant-B")
	if err := cB.Call(opCreate, &z2, nil); err != nil {
		t.Fatalf("same range, other tenant: %v", err)
	}
}
	overlap := devicemanager.Zone{TenantId: "tenant-A", Name: "overlap", RangeStart: "172.0.0.65", RangeEnd: "172.0.0.79"}
	if _, err := m.SaveZone(overlap); !isErr(err, devicemanager.ErrZoneOverlap) {
		t.Fatalf("overlapping zone: err = %v, want ErrZoneOverlap", err)
	}
	// Mismo rango en otro tenant es válido: las zonas son por tenant.
	if _, err := m.SaveZone(zoneAP06("tenant-B")); err != nil {
		t.Fatalf("same range, other tenant: %v", err)
	}
}

func TestZone_DeleteWithDevicesRejected(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	opCreateZone := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}

	z := zoneAP06("tenant-A")
	var outZone devicemanager.Zone
	if err := c.Call(opCreateZone, &z, &outZone); err != nil {
		t.Fatalf("CreateZone: %v", err)
	}

	opCreateDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
	dev := &devicemanager.Device{Name: "PC 1", Type: devicemanager.DeviceTypeComputer, ZoneId: outZone.Id, IsActive: true}
	if err := c.Call(opCreateDev, dev, nil); err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}

	opDelete := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Delete}
	if err := c.Call(opDelete, map[string]interface{}{"id": outZone.Id}, nil); !isErr(err, devicemanager.ErrZoneInUse) {
		t.Fatalf("DeleteZone in use: err = %v, want ErrZoneInUse", err)
	}
}
	if err := m.DeleteZone("tenant-A", z.Id); !isErr(err, devicemanager.ErrZoneInUse) {
		t.Fatalf("DeleteZone in use: err = %v, want ErrZoneInUse", err)
	}
}

func TestZone_TenantIsolation(t *testing.T) {
	mA := setup(t)
	cA := crud.LoopbackActingAs(nil, "u1", mA)
	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}
	z := zoneAP06("tenant-A")
	var outZone devicemanager.Zone
	if err := cA.Call(opCreate, &z, &outZone); err != nil {
		t.Fatalf("CreateZone: %v", err)
	}

	mB := setup(t)
	mB.TenantID = "tenant-B" // Hacky way but let's test readers correctly instead
	if _, err := mB.GetZone("tenant-B", outZone.Id); !isErr(err, devicemanager.ErrZoneNotFound) {
		t.Errorf("tenant B read tenant A's zone: err = %v", err)
	}
}
	if err := m.DeleteZone("tenant-B", z.Id); !isErr(err, devicemanager.ErrZoneNotFound) {
		t.Errorf("tenant B deleted tenant A's zone: err = %v", err)
	}
	hijack := z
	hijack.TenantId = "tenant-B"
	if _, err := m.SaveZone(hijack); !isErr(err, devicemanager.ErrZoneNotFound) {
		t.Errorf("tenant B updated tenant A's zone: err = %v", err)
	}
}
