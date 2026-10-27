package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/model"
)

func TestRowsDeclaredOnFields(t *testing.T) {
	d := devicemanager.Device{
		Id:       "dev_1",
		Name:     "Pc Recepcion",
		Location: "Room 1",
		Access:   "full",
	}
	rD := model.RowOf(&d)
	if rD.Title != "Pc Recepcion" {
		t.Errorf("Device Title = %q, want %q", rD.Title, "Pc Recepcion")
	}
	if rD.Detail != "Room 1 · full" {
		t.Errorf("Device Detail = %q, want %q", rD.Detail, "Room 1 · full")
	}

	z := devicemanager.Zone{
		Id:         "zone_1",
		Name:       "Reception",
		RangeStart: "10.0.0.1",
		RangeEnd:   "10.0.0.254",
	}
	rZ := model.RowOf(&z)
	if rZ.Title != "Reception" {
		t.Errorf("Zone Title = %q, want %q", rZ.Title, "Reception")
	}
	if rZ.Detail != "10.0.0.1 · 10.0.0.254" {
		t.Errorf("Zone Detail = %q, want %q", rZ.Detail, "10.0.0.1 · 10.0.0.254")
	}

	ni := devicemanager.NetworkInterface{
		Id:    "ni_1",
		Label: "lan",
		Ip:    "192.168.1.100",
		Mac:   "00:11:22:33:44:55",
	}
	rNI := model.RowOf(&ni)
	if rNI.Title != "lan" {
		t.Errorf("NetworkInterface Title = %q, want %q", rNI.Title, "lan")
	}
	if rNI.Detail != "00:11:22:33:44:55 · 192.168.1.100" {
		t.Errorf("NetworkInterface Detail = %q, want %q", rNI.Detail, "00:11:22:33:44:55 · 192.168.1.100")
	}
}
