package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/network"
)

// deviceInZone creates a device in a fresh AP-06 zone (.60–.69).
func deviceInZone(t *testing.T, m *devicemanager.Module) (devicemanager.Device, devicemanager.Zone) {
	t.Helper()
	z, err := m.SaveZone(zoneAP06("tenant-A"))
	if err != nil {
		t.Fatalf("SaveZone: %v", err)
	}
	d, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "PC Ecografia 1", Type: devicemanager.DeviceTypeComputer,
		ZoneId: z.Id, Access: network.AccessInternetName, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	return d, z
}

func iface(d devicemanager.Device, label, mac, ip string) devicemanager.NetworkInterface {
	return devicemanager.NetworkInterface{TenantId: d.TenantId, DeviceId: d.Id, Label: label, Mac: mac, Ip: ip}
}

func TestInterface_AutoAssignsLowestFreeIP(t *testing.T) {
	m := setup(t)
	d, _ := deviceInZone(t, m)

	first, err := m.SaveNetworkInterface(iface(d, "wifi", "", ""))
	if err != nil || first.Ip != "172.0.0.60" {
		t.Fatalf("first auto IP = %q (err %v), want 172.0.0.60", first.Ip, err)
	}
	if _, err := m.SaveNetworkInterface(iface(d, "manual", "", "172.0.0.62")); err != nil {
		t.Fatalf("explicit .62: %v", err)
	}
	second, err := m.SaveNetworkInterface(iface(d, "ethernet", "", ""))
	if err != nil || second.Ip != "172.0.0.61" {
		t.Fatalf("second auto IP = %q (err %v), want 172.0.0.61", second.Ip, err)
	}
	third, err := m.SaveNetworkInterface(iface(d, "usb", "", ""))
	if err != nil || third.Ip != "172.0.0.63" {
		t.Fatalf("third auto IP = %q (err %v), want 172.0.0.63 (skipping the hand-set .62)", third.Ip, err)
	}
}

func TestInterface_ZoneFull(t *testing.T) {
	m := setup(t)
	z, _ := m.SaveZone(devicemanager.Zone{TenantId: "tenant-A", Name: "tiny", RangeStart: "10.0.0.1", RangeEnd: "10.0.0.1"})
	d, _ := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "PC", Type: devicemanager.DeviceTypeComputer, ZoneId: z.Id, IsActive: true})
	if _, err := m.SaveNetworkInterface(iface(d, "lan a", "", "")); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := m.SaveNetworkInterface(iface(d, "lan b", "", "")); !isErr(err, devicemanager.ErrZoneFull) {
		t.Fatalf("second in a full zone: err = %v, want ErrZoneFull", err)
	}
}

func TestInterface_NoZoneNeedsExplicitIP(t *testing.T) {
	m := setup(t)
	d, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Servidor", Type: devicemanager.DeviceTypeServer, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if _, err := m.SaveNetworkInterface(iface(d, "lan", "", "")); !isValidation(err, devicemanager.ErrNoZone) {
		t.Fatalf("empty IP without zone: err = %v, want ValidationError{ErrNoZone}", err)
	}
	// Without a zone any IP is accepted — e.g. the server's loopback aliases used for login.
	for _, ip := range []string{"127.0.0.1", "192.168.50.5"} {
		if _, err := m.SaveNetworkInterface(iface(d, "lan "+ip, "", ip)); err != nil {
			t.Errorf("IP %s on a device without zone: %v", ip, err)
		}
	}
}

func TestInterface_IPOutsideZone(t *testing.T) {
	m := setup(t)
	d, _ := deviceInZone(t, m)
	if _, err := m.SaveNetworkInterface(iface(d, "wifi", "", "172.0.0.47")); !isValidation(err, devicemanager.ErrIPOutsideZone) {
		t.Fatalf(".47 in zone .60–.69: err = %v, want ValidationError{ErrIPOutsideZone}", err)
	}
}

func TestInterface_MACRules(t *testing.T) {
	m := setup(t)
	d, _ := deviceInZone(t, m)

	ni, err := m.SaveNetworkInterface(iface(d, "wifi", "48-f1-7f-d9-d7-b7", ""))
	if err != nil {
		t.Fatalf("SaveNetworkInterface: %v", err)
	}
	if ni.Mac != "48:F1:7F:D9:D7:B7" {
		t.Errorf("stored MAC = %q, want canonical 48:F1:7F:D9:D7:B7", ni.Mac)
	}
	// The incident of 2026-10-07: Windows "change daily" invented this one.
	if _, err := m.SaveNetworkInterface(iface(d, "wifi2", "4A:C5:93:7A:12:DE", "")); !isValidation(err, devicemanager.ErrRandomizedMAC) {
		t.Errorf("randomized MAC: err = %v, want ValidationError{ErrRandomizedMAC}", err)
	}
	if _, err := m.SaveNetworkInterface(iface(d, "dup", "48:F1:7F:D9:D7:B7", "")); !isErr(err, devicemanager.ErrMACAlreadyExists) {
		t.Errorf("duplicate MAC: err = %v, want ErrMACAlreadyExists", err)
	}
	// Updating the same interface with its own MAC is not a duplicate.
	ni.Label = "wifi intel"
	if _, err := m.SaveNetworkInterface(ni); err != nil {
		t.Errorf("update keeping its own MAC: %v", err)
	}
}

func TestInterface_DeleteDeviceDeletesInterfaces(t *testing.T) {
	m := setup(t)
	d, _ := deviceInZone(t, m)
	if _, err := m.SaveNetworkInterface(iface(d, "wifi", "", "")); err != nil {
		t.Fatalf("SaveNetworkInterface: %v", err)
	}
	if err := m.DeleteDevice("tenant-A", d.Id); err != nil {
		t.Fatalf("DeleteDevice: %v", err)
	}
	left, err := m.ListNetworkInterfaces("tenant-A", d.Id)
	if err != nil || len(left) != 0 {
		t.Fatalf("interfaces after deleting the device = %+v (err %v), want none", left, err)
	}
}

func TestInterface_FindByIPThroughInterface(t *testing.T) {
	m := setup(t)
	d, _ := deviceInZone(t, m)
	if _, err := m.SaveNetworkInterface(iface(d, "wifi", "", "172.0.0.61")); err != nil {
		t.Fatalf("SaveNetworkInterface: %v", err)
	}
	got, err := m.FindByIP("tenant-A", "172.0.0.61")
	if err != nil || got.Id != d.Id {
		t.Fatalf("FindByIP = %+v (err %v), want device %s", got, err, d.Id)
	}
}

// isValidation reports whether err is a ValidationError wrapping want.
func isValidation(err, want error) bool {
	v, ok := err.(devicemanager.ValidationError)
	return ok && v.Err == want
}
