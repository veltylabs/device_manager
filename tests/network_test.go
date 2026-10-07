package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/form"
	"webtyp.com/network"
)

func TestNetworkHosts_OnlyActiveDevicesInterfacesWithMAC(t *testing.T) {
	m := setup(t)
	d, _ := deviceInZone(t, m)
	if _, err := m.SaveNetworkInterface(iface(d, "wifi", "48:F1:7F:D9:D7:B7", "172.0.0.61")); err != nil {
		t.Fatalf("wifi: %v", err)
	}
	if _, err := m.SaveNetworkInterface(iface(d, "sin mac", "", "172.0.0.62")); err != nil {
		t.Fatalf("interface without MAC: %v", err)
	}
	off, _ := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Retired", Type: devicemanager.DeviceTypeComputer, ZoneId: d.ZoneId, IsActive: true})
	if _, err := m.SaveNetworkInterface(iface(off, "wifi", "00:1A:2B:00:09:99", "")); err != nil {
		t.Fatalf("retired device iface: %v", err)
	}
	if err := m.DeactivateDevice("tenant-A", off.Id); err != nil {
		t.Fatalf("DeactivateDevice: %v", err)
	}

	hosts, err := devicemanager.NetworkHosts{Devices: m, TenantID: "tenant-A"}.Hosts()
	if err != nil {
		t.Fatalf("Hosts: %v", err)
	}
	want := network.Host{Name: "PC Ecografia 1 (wifi)", MAC: "48:F1:7F:D9:D7:B7", IP: "172.0.0.61", Access: network.AccessInternet}
	if len(hosts) != 1 || hosts[0] != want {
		t.Fatalf("Hosts = %+v, want exactly %+v", hosts, want)
	}
}

func TestNetworkHosts_ImportHosts(t *testing.T) {
	m := setup(t)
	z, _ := m.SaveZone(zoneAP06("tenant-A"))
	existing, _ := deviceInOtherZone(t, m)
	if _, err := m.SaveNetworkInterface(iface(existing, "lan", "00:1A:2B:00:00:01", "10.10.0.5")); err != nil {
		t.Fatalf("existing interface: %v", err)
	}

	found := []network.Discovered{
		{MAC: "48:F1:7F:D9:D7:B7", IP: "172.0.0.61", Name: "INTERNET PC-GASTRO-NEW", Internet: true}, // ok, internet, in zone
		{MAC: "00:1A:2B:00:00:02", IP: "172.0.0.62"},                                                 // ok, local, unnamed
		{IP: "172.0.0.63", Name: "no mac"},                                                           // skipped: no MAC
		{MAC: "00:1A:2B:00:00:03", Name: "rule only", Internet: true},                                // skipped: no IP
		{MAC: "4A:C5:93:7A:12:DE", IP: "172.0.0.64"},                                                 // skipped: randomized
		{MAC: "00:1A:2B:00:00:01", IP: "10.10.0.5"},                                                  // skipped: registered
	}
	res, err := devicemanager.NetworkHosts{Devices: m, TenantID: "tenant-A"}.ImportHosts(found)
	if err != nil {
		t.Fatalf("ImportHosts: %v", err)
	}
	if res.Created != 2 {
		t.Errorf("Created = %d, want 2", res.Created)
	}
	wantReasons := []string{devicemanager.SkipNoMAC, devicemanager.SkipNoIP, devicemanager.SkipRandomizedMAC, devicemanager.SkipAlreadyRegistered}
	if len(res.Skipped) != len(wantReasons) {
		t.Fatalf("Skipped = %+v, want %d entries", res.Skipped, len(wantReasons))
	}
	for i, r := range wantReasons {
		if res.Skipped[i].Reason != r {
			t.Errorf("Skipped[%d].Reason = %q, want %q", i, res.Skipped[i].Reason, r)
		}
	}

	imported, err := m.FindByIP("tenant-A", "172.0.0.61")
	if err != nil {
		t.Fatalf("imported device not found by IP: %v", err)
	}
	if imported.Name != "INTERNET PC-GASTRO-NEW" || imported.Access != network.AccessInternetName || imported.ZoneId != z.Id {
		t.Errorf("imported device = %+v, want name from the router comment, access internet, zone %s", imported, z.Id)
	}
	unnamed, err := m.FindByIP("tenant-A", "172.0.0.62")
	if err != nil || unnamed.Access != network.AccessLocalName || unnamed.Name != "Imported 00-1A-2B-00-00-02" {
		t.Errorf("unnamed import = %+v (err %v), want local access and name %q", unnamed, err, "Imported 00-1A-2B-00-00-02")
	}
}

// deviceInOtherZone creates a device in a zone that does not overlap AP-06.
func deviceInOtherZone(t *testing.T, m *devicemanager.Module) (devicemanager.Device, devicemanager.Zone) {
	t.Helper()
	z, err := m.SaveZone(devicemanager.Zone{TenantId: "tenant-A", Name: "Servidores", RangeStart: "10.10.0.1", RangeEnd: "10.10.0.9"})
	if err != nil {
		t.Fatalf("SaveZone: %v", err)
	}
	d, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Servidor", Type: devicemanager.DeviceTypeServer, ZoneId: z.Id, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	return d, z
}

// The forms crudview builds must show exactly the editable fields.
func TestFormWidgets(t *testing.T) {
	ids := &mockIDGen{}
	cases := []struct {
		record   interface{}
		want     []string
		mustLack []string
	}{
		{&devicemanager.Device{}, []string{"name", "type", "location", "zone_id", "access", "is_active"}, []string{"ip", "tenant_id"}},
		{&devicemanager.NetworkInterface{}, []string{"device_id", "label", "mac", "ip"}, []string{"tenant_id"}},
		{&devicemanager.Zone{}, []string{"name", "range_start", "range_end"}, []string{"tenant_id"}},
	}
	for _, c := range cases {
		var f *form.Form
		var err error
		switch r := c.record.(type) {
		case *devicemanager.Device:
			f, err = form.New("dm", r, ids)
		case *devicemanager.NetworkInterface:
			f, err = form.New("dm", r, ids)
		case *devicemanager.Zone:
			f, err = form.New("dm", r, ids)
		}
		if err != nil {
			t.Fatalf("form.New(%T): %v", c.record, err)
		}
		for _, name := range c.want {
			if f.Input(name) == nil {
				t.Errorf("%T form lacks input %q", c.record, name)
			}
		}
		for _, name := range c.mustLack {
			if f.Input(name) != nil {
				t.Errorf("%T form has input %q, which must not be editable", c.record, name)
			}
		}
	}
}
