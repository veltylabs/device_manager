package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/view"
)

func TestScreenSave(t *testing.T) {
	dm := setup(t)

	caller := view.NewRecords().Caller(crud.LoopbackActingAs(nil, "user-1", dm))

	// Device
	dlister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName}, func() model.TableSlice { return &devicemanager.DeviceList{} })
	saver := dlister.(view.Saver)
	deleter := dlister.(view.Deleter)

	d := &devicemanager.Device{Name: "test-device", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true}
	var saveErr error
	saver.Save([]model.Model{d}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if d.Id == "" {
		list, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{Limit: int64(crud.Cap(0))})
		d.Id = list[0].Id
	}

	var delErr error
	deleter.Delete([]string{d.Id}, func(err error) { delErr = err })
	if delErr != nil {
		t.Fatal(delErr)
	}

	// Zone
	zlister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName}, func() model.TableSlice { return &devicemanager.ZoneList{} })
	zsaver := zlister.(view.Saver)
	zdeleter := zlister.(view.Deleter)

	z := &devicemanager.Zone{Name: "test-zone", RangeStart: "10.0.0.1", RangeEnd: "10.0.0.10"}
	zsaver.Save([]model.Model{z}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if z.Id == "" {
		list, _ := dm.ListZones("tenant-A", "", "", 0)
		z.Id = list[len(list)-1].Id
	}

	// NetworkInterface
	nilister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName}, func() model.TableSlice { return &devicemanager.NetworkInterfaceList{} })
	nisaver := nilister.(view.Saver)
	nideleter := nilister.(view.Deleter)

	ni := &devicemanager.NetworkInterface{DeviceId: "test", Label: "eth0", Mac: "00:11:22:33:44:55", Ip: "10.0.0.1"}
	nisaver.Save([]model.Model{ni}, func(err error) { saveErr = err })
	if saveErr == nil {
		t.Fatal("expected error saving interface with invalid device id")
	}

	// Create device first for the interface
	devForIface := &devicemanager.Device{Name: "dev2", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true}
	saver.Save([]model.Model{devForIface}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	list_dev2, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{Limit: int64(crud.Cap(0))})
	for _, row := range list_dev2 { // ids sort as text: pick the device by its name
		if row.Name == devForIface.Name {
			devForIface.Id = row.Id
		}
	}

	ni.DeviceId = devForIface.Id
	nisaver.Save([]model.Model{ni}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if ni.Id == "" {
		list, _ := dm.ListNetworkInterfaces("tenant-A", ni.DeviceId, "", "", 0)
		ni.Id = list[0].Id
	}

	// IP Conflict test
	z2 := &devicemanager.Zone{Name: "test-zone-2", RangeStart: "10.10.0.1", RangeEnd: "10.10.0.1"}
	zsaver.Save([]model.Model{z2}, func(err error) {})
	list_z2, _ := dm.ListZones("tenant-A", "", "", 0)
	for _, row := range list_z2 { // ids sort as text ("test-id-9" > "test-id-12"): pick the zone by its name
		if row.Name == z2.Name {
			z2.Id = row.Id
		}
	}

	devForConflict := &devicemanager.Device{Name: "dev-conflict", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true, ZoneId: z2.Id}
	saver.Save([]model.Model{devForConflict}, func(err error) {})
	list_devConf, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{Limit: int64(crud.Cap(0))})
	for _, row := range list_devConf { // ids sort as text: pick the device by its name
		if row.Name == devForConflict.Name {
			devForConflict.Id = row.Id
		}
	}

	ni2 := &devicemanager.NetworkInterface{DeviceId: devForConflict.Id, Label: "eth0"}
	nisaver.Save([]model.Model{ni2}, func(err error) {}) // Takes 10.10.0.1

	devForConflict2 := &devicemanager.Device{Name: "dev-conflict2", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true, ZoneId: z2.Id}
	saver.Save([]model.Model{devForConflict2}, func(err error) {})
	list_devConf2, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{Limit: int64(crud.Cap(0))})
	for _, row := range list_devConf2 { // ids sort as text: pick the device by its name
		if row.Name == devForConflict2.Name {
			devForConflict2.Id = row.Id
		}
	}

	ni3 := &devicemanager.NetworkInterface{DeviceId: devForConflict2.Id, Label: "eth1"}
	nisaver.Save([]model.Model{ni3}, func(err error) { saveErr = err })
	if model.KindOf(saveErr) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict (ZoneFull), got %v", saveErr)
	}

	// Attempt manual IP conflict
	ni4 := &devicemanager.NetworkInterface{DeviceId: devForConflict2.Id, Label: "eth2", Ip: "10.10.0.1"}
	nisaver.Save([]model.Model{ni4}, func(err error) { saveErr = err })
	if model.KindOf(saveErr) != model.ErrorConflict {
		t.Fatalf("expected ErrorConflict (IPAlreadyExists), got %v", saveErr)
	}

	// Cleanup
	nideleter.Delete([]string{ni.Id}, func(err error) { delErr = err })
	if delErr != nil {
		t.Fatal(delErr)
	}
	zdeleter.Delete([]string{z.Id}, func(err error) { delErr = err })
	if delErr != nil {
		t.Fatal(delErr)
	}
}

func TestScreen_TenantFromRequestIsRejected(t *testing.T) {
	dm := setup(t)

	caller := view.NewRecords().Caller(crud.LoopbackActingAs(nil, "user-1", dm))

	dlister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName}, func() model.TableSlice { return &devicemanager.DeviceList{} })
	saver := dlister.(view.Saver)

	d := &devicemanager.Device{TenantId: "other-tenant", Name: "test-device", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true}
	var saveErr error
	saver.Save([]model.Model{d}, func(err error) { saveErr = err })
	if saveErr == nil {
		t.Fatal("a tenant_id sent by the client must be refused (ServerOnly)")
	}
	if list, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{Limit: int64(crud.Cap(0))}); len(list) != 0 {
		t.Fatalf("nothing may be stored, got %d devices", len(list))
	}

	zlister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName}, func() model.TableSlice { return &devicemanager.ZoneList{} })
	zsaver := zlister.(view.Saver)

	z := &devicemanager.Zone{TenantId: "other-tenant", Name: "test-zone", RangeStart: "10.0.0.1", RangeEnd: "10.0.0.10"}
	zsaver.Save([]model.Model{z}, func(err error) { saveErr = err })
	if saveErr == nil {
		t.Fatal("a tenant_id sent by the client must be refused (ServerOnly)")
	}
	if list, _ := dm.ListZones("tenant-A", "", "", 0); len(list) != 0 {
		t.Fatalf("nothing may be stored, got %d zones", len(list))
	}
}
