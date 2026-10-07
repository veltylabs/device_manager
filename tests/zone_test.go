package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
)

func zoneAP06(tenant string) devicemanager.Zone {
	return devicemanager.Zone{TenantId: tenant, Name: "AP-06 piso 2", RangeStart: "172.0.0.60", RangeEnd: "172.0.0.69"}
}

func TestZone_SaveListDelete(t *testing.T) {
	m := setup(t)
	z, err := m.SaveZone(zoneAP06("tenant-A"))
	if err != nil {
		t.Fatalf("SaveZone: %v", err)
	}
	zones, err := m.ListZones("tenant-A")
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
	if _, err := m.GetZone("tenant-A", z.Id); err != devicemanager.ErrZoneNotFound {
		t.Fatalf("GetZone after delete: err = %v, want ErrZoneNotFound", err)
	}
}

func TestZone_InvalidRange(t *testing.T) {
	m := setup(t)
	for _, z := range []devicemanager.Zone{
		{TenantId: "tenant-A", Name: "reversed", RangeStart: "172.0.0.69", RangeEnd: "172.0.0.60"},
		{TenantId: "tenant-A", Name: "ipv6", RangeStart: "::1", RangeEnd: "::2"},
	} {
		if _, err := m.SaveZone(z); err == nil {
			t.Errorf("SaveZone(%s) = nil, want a ValidationError", z.Name)
		} else if _, ok := err.(devicemanager.ValidationError); !ok {
			t.Errorf("SaveZone(%s) = %v (%T), want ValidationError", z.Name, err, err)
		}
	}
}

func TestZone_Overlap(t *testing.T) {
	m := setup(t)
	if _, err := m.SaveZone(zoneAP06("tenant-A")); err != nil {
		t.Fatalf("SaveZone: %v", err)
	}
	overlap := devicemanager.Zone{TenantId: "tenant-A", Name: "overlap", RangeStart: "172.0.0.65", RangeEnd: "172.0.0.79"}
	if _, err := m.SaveZone(overlap); err != devicemanager.ErrZoneOverlap {
		t.Fatalf("overlapping zone: err = %v, want ErrZoneOverlap", err)
	}
	// Same range in another tenant is fine: zones are per tenant.
	if _, err := m.SaveZone(zoneAP06("tenant-B")); err != nil {
		t.Fatalf("same range, other tenant: %v", err)
	}
}

func TestZone_DeleteInUse(t *testing.T) {
	m := setup(t)
	z, _ := m.SaveZone(zoneAP06("tenant-A"))
	if _, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "PC 1", Type: devicemanager.DeviceTypeComputer, ZoneId: z.Id, IsActive: true}); err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if err := m.DeleteZone("tenant-A", z.Id); err != devicemanager.ErrZoneInUse {
		t.Fatalf("DeleteZone in use: err = %v, want ErrZoneInUse", err)
	}
}

func TestZone_TenantIsolation(t *testing.T) {
	m := setup(t)
	z, _ := m.SaveZone(zoneAP06("tenant-A"))
	if _, err := m.GetZone("tenant-B", z.Id); err != devicemanager.ErrZoneNotFound {
		t.Errorf("tenant B read tenant A's zone: err = %v", err)
	}
	if err := m.DeleteZone("tenant-B", z.Id); err != devicemanager.ErrZoneNotFound {
		t.Errorf("tenant B deleted tenant A's zone: err = %v", err)
	}
	hijack := z
	hijack.TenantId = "tenant-B"
	if _, err := m.SaveZone(hijack); err != devicemanager.ErrZoneNotFound {
		t.Errorf("tenant B updated tenant A's zone: err = %v", err)
	}
}
