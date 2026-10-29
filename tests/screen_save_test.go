package tests

import (
	"strings"
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/model"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

func TestScreenSave(t *testing.T) {
	dm := setup(t)

	caller := view.NewRecords().Caller(loopback.ActingAs("tenant-A", "user-1", dm))

	// Device
	dlister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName, List: devicemanager.OpListDevices, Save: devicemanager.OpUpsertDevice, Delete: devicemanager.OpDeleteDevice}, func() model.ModelSlice { return &devicemanager.DeviceList{} })
	saver := dlister.(view.Saver)
	deleter := dlister.(view.Deleter)

	d := &devicemanager.Device{TenantId: "tenant-A", Name: "test-device", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true}
	var saveErr error
	saver.Save([]model.Model{d}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if d.Id == "" {
		list, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{})
		d.Id = list[0].Id
	}

	var delErr error
	deleter.Delete([]string{d.Id}, func(err error) { delErr = err })
	if delErr != nil {
		t.Fatal(delErr)
	}

	// Zone
	zlister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName, List: devicemanager.OpListZones, Save: devicemanager.OpSaveZone, Delete: devicemanager.OpDeleteZone}, func() model.ModelSlice { return &devicemanager.ZoneList{} })
	zsaver := zlister.(view.Saver)
	zdeleter := zlister.(view.Deleter)

	z := &devicemanager.Zone{TenantId: "tenant-A", Name: "test-zone", RangeStart: "10.0.0.1", RangeEnd: "10.0.0.10"}
	zsaver.Save([]model.Model{z}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if z.Id == "" {
		list, _ := dm.ListZones("tenant-A")
		z.Id = list[len(list)-1].Id
	}

	// NetworkInterface
	nilister := view.NewCallerLister(caller, view.Ops{Module: devicemanager.ModelName, List: devicemanager.OpListNetworkInterfaces, Save: devicemanager.OpSaveNetworkInterface, Delete: devicemanager.OpDeleteNetworkInterface}, func() model.ModelSlice { return &devicemanager.NetworkInterfaceList{} })
	nisaver := nilister.(view.Saver)
	nideleter := nilister.(view.Deleter)

	ni := &devicemanager.NetworkInterface{TenantId: "tenant-A", DeviceId: "test", Label: "eth0", Mac: "00:11:22:33:44:55", Ip: "10.0.0.1"}
	nisaver.Save([]model.Model{ni}, func(err error) { saveErr = err })
	if saveErr == nil {
		t.Fatal("expected error saving interface with invalid device id")
	}

	// Create device first for the interface
	devForIface := &devicemanager.Device{TenantId: "tenant-A", Name: "dev2", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true}
	saver.Save([]model.Model{devForIface}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	list_dev2, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{})
	devForIface.Id = list_dev2[len(list_dev2)-1].Id

	ni.DeviceId = devForIface.Id
	nisaver.Save([]model.Model{ni}, func(err error) { saveErr = err })
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if ni.Id == "" {
		list, _ := dm.ListNetworkInterfaces("tenant-A", ni.DeviceId)
		ni.Id = list[0].Id
	}

	// IP Conflict test
	z2 := &devicemanager.Zone{TenantId: "tenant-A", Name: "test-zone-2", RangeStart: "10.10.0.1", RangeEnd: "10.10.0.1"}
	zsaver.Save([]model.Model{z2}, func(err error) {})
	list_z2, _ := dm.ListZones("tenant-A")
	z2.Id = list_z2[len(list_z2)-1].Id

	devForConflict := &devicemanager.Device{TenantId: "tenant-A", Name: "dev-conflict", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true, ZoneId: z2.Id}
	saver.Save([]model.Model{devForConflict}, func(err error) {})
	list_devConf, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{})
	devForConflict.Id = list_devConf[len(list_devConf)-1].Id

	ni2 := &devicemanager.NetworkInterface{TenantId: "tenant-A", DeviceId: devForConflict.Id, Label: "eth0"}
	nisaver.Save([]model.Model{ni2}, func(err error) {}) // Takes 10.10.0.1

	devForConflict2 := &devicemanager.Device{TenantId: "tenant-A", Name: "dev-conflict2", Type: devicemanager.DeviceTypeComputer, Access: "local", IsActive: true, ZoneId: z2.Id}
	saver.Save([]model.Model{devForConflict2}, func(err error) {})
	list_devConf2, _ := dm.ListDevices("tenant-A", devicemanager.DeviceFilter{})
	devForConflict2.Id = list_devConf2[len(list_devConf2)-1].Id

	ni3 := &devicemanager.NetworkInterface{TenantId: "tenant-A", DeviceId: devForConflict2.Id, Label: "eth1"}
	nisaver.Save([]model.Model{ni3}, func(err error) { saveErr = err })
	if !strings.Contains(saveErr.Error(), `"status":409`) {
		t.Fatalf("expected ErrorConflict (ZoneFull), got %v", saveErr)
	}

	// Attempt manual IP conflict
	ni4 := &devicemanager.NetworkInterface{TenantId: "tenant-A", DeviceId: devForConflict2.Id, Label: "eth2", Ip: "10.10.0.1"}
	nisaver.Save([]model.Model{ni4}, func(err error) { saveErr = err })
	if !strings.Contains(saveErr.Error(), `"status":409`) {
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
