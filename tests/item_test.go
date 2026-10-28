package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
)

func TestItemsDeclaredOnFields(t *testing.T) {
	d := devicemanager.Device{
		Id:       "dev_1",
		Name:     "Pc Recepcion",
		Location: "Room 1",
		Access:   "full",
	}
	rD := d.Item()
	if rD.Label != "Pc Recepcion" {
		t.Errorf("Device Title = %q, want %q", rD.Label, "Pc Recepcion")
	}
	if rD.Description != "Room 1 · full" {
		t.Errorf("Device Detail = %q, want %q", rD.Description, "Room 1 · full")
	}

	z := devicemanager.Zone{
		Id:         "zone_1",
		Name:       "Reception",
		RangeStart: "10.0.0.1",
		RangeEnd:   "10.0.0.254",
	}
	rZ := z.Item()
	if rZ.Label != "Reception" {
		t.Errorf("Zone Title = %q, want %q", rZ.Label, "Reception")
	}
	if rZ.Description != "10.0.0.1 · 10.0.0.254" {
		t.Errorf("Zone Detail = %q, want %q", rZ.Description, "10.0.0.1 · 10.0.0.254")
	}

	ni := devicemanager.NetworkInterface{
		Id:    "ni_1",
		Label: "lan",
		Ip:    "192.168.1.100",
		Mac:   "00:11:22:33:44:55",
	}
	rNI := ni.Item()
	if rNI.Label != "lan" {
		t.Errorf("NetworkInterface Title = %q, want %q", rNI.Label, "lan")
	}
	if rNI.Description != "00:11:22:33:44:55 · 192.168.1.100" {
		t.Errorf("NetworkInterface Detail = %q, want %q", rNI.Description, "00:11:22:33:44:55 · 192.168.1.100")
	}
}
