package tests

import (
	"webtyp.com/model"
	"testing"
	"webtyp.com/crud"
	"webtyp.com/router"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/network"
)

func deviceInZone(t *testing.T, c router.Caller) (devicemanager.Device, devicemanager.Zone) {
	t.Helper()
	opCreateZone := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}
	z := &devicemanager.Zone{Name: "AP-06 piso 2", RangeStart: "172.0.0.60", RangeEnd: "172.0.0.69"}
	var outZoneRes crud.CreateResult
	if err := callSync(c, opCreateZone, &crud.CreateArgs{Records: []model.Model{z}}, &outZoneRes); err != nil {
		t.Fatalf("CreateZone: %v", err)
	}
	zId := outZoneRes.IDs[0]
	z.Id = zId

	opCreateDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
	d := &devicemanager.Device{Name: "PC Ecografia 1", Type: devicemanager.DeviceTypeComputer, ZoneId: zId, Access: network.AccessInternetName, IsActive: true}
	var outDevRes crud.CreateResult
	if err := callSync(c, opCreateDev, &crud.CreateArgs{Records: []model.Model{d}}, &outDevRes); err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	devId := outDevRes.IDs[0]
	d.Id = devId
	d.TenantId = "tenant-A"
	z.TenantId = "tenant-A"
	return *d, *z
}

func iface(d devicemanager.Device, label, mac, ip string) devicemanager.NetworkInterface {
	return devicemanager.NetworkInterface{TenantId: d.TenantId, DeviceId: d.Id, Label: label, Mac: mac, Ip: ip}
}

func TestInterface_AutoAssignsLowestFreeIP(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	d, _ := deviceInZone(t, c)
	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}

	ni1 := iface(d, "wifi", "", "")
	var first crud.CreateResult
	err := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni1}}, &first)
	if err != nil {
		t.Fatalf("first auto IP: %v", err)
	}

	ni2 := iface(d, "manual", "", "172.0.0.62")
	if err := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni2}}, nil); err != nil {
		t.Fatalf("explicit .62: %v", err)
	}

	ni3 := iface(d, "ethernet", "", "")
	var second crud.CreateResult
	if err := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni3}}, &second); err != nil {
		t.Fatalf("second auto IP: %v", err)
	}

	ni4 := iface(d, "usb", "", "")
	var third crud.CreateResult
	if err := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni4}}, &third); err != nil {
		t.Fatalf("third auto IP: %v", err)
	}
}

func TestInterface_ZoneFull(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)

	opCreateZone := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}
	z := &devicemanager.Zone{Name: "tiny", RangeStart: "10.0.0.1", RangeEnd: "10.0.0.1"}
	var outZone crud.CreateResult
	callSync(c, opCreateZone, &crud.CreateArgs{Records: []model.Model{z}}, &outZone)
	zId := outZone.IDs[0]

	opCreateDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
	d := &devicemanager.Device{Name: "PC", Type: devicemanager.DeviceTypeComputer, ZoneId: zId, IsActive: true}
	var outDev crud.CreateResult
	callSync(c, opCreateDev, &crud.CreateArgs{Records: []model.Model{d}}, &outDev)
	dId := outDev.IDs[0]

	opCreateNi := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni1 := iface(devicemanager.Device{Id: dId, TenantId: "tenant-A"}, "lan a", "", "")
	if err := callSync(c, opCreateNi, &crud.CreateArgs{Records: []model.Model{&ni1}}, nil); err != nil {
		t.Fatalf("first: %v", err)
	}

	ni2 := iface(devicemanager.Device{Id: dId, TenantId: "tenant-A"}, "lan b", "", "")
	err := callSync(c, opCreateNi, &crud.CreateArgs{Records: []model.Model{&ni2}}, nil)
	if !isErr(err, devicemanager.ErrZoneFull) {
		t.Fatalf("second in a full zone: err = %v, want ErrZoneFull", err)
	}
}

func TestInterface_NoZoneNeedsExplicitIP(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)

	opCreateDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
	d := &devicemanager.Device{Name: "Servidor", Type: devicemanager.DeviceTypeServer, IsActive: true}
	var outDev crud.CreateResult
	if err := callSync(c, opCreateDev, &crud.CreateArgs{Records: []model.Model{d}}, &outDev); err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	dId := outDev.IDs[0]

	opCreateNi := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni1 := iface(devicemanager.Device{Id: dId, TenantId: "tenant-A"}, "lan", "", "")
	err := callSync(c, opCreateNi, &crud.CreateArgs{Records: []model.Model{&ni1}}, nil)
	if !isValidation(err, devicemanager.ErrNoZone) {
		t.Fatalf("empty IP without zone: err = %v, want ValidationError{ErrNoZone}", err)
	}

	for _, ip := range []string{"127.0.0.1", "192.168.50.5"} {
		ni := iface(devicemanager.Device{Id: dId, TenantId: "tenant-A"}, "lan "+ip, "", ip)
		if err := callSync(c, opCreateNi, &crud.CreateArgs{Records: []model.Model{&ni}}, nil); err != nil {
			t.Errorf("IP %s on a device without zone: %v", ip, err)
		}
	}
}

func TestInterface_IPOutsideZone(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	d, _ := deviceInZone(t, c)

	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni := iface(d, "wifi", "", "172.0.0.47")
	err := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni}}, nil)
	if !isValidation(err, devicemanager.ErrIPOutsideZone) {
		t.Fatalf(".47 in zone .60–.69: err = %v, want ValidationError{ErrIPOutsideZone}", err)
	}
}

func TestInterface_MACRules(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	d, _ := deviceInZone(t, c)

	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni1 := iface(d, "wifi", "48-f1-7f-d9-d7-b7", "")
	var outNi crud.CreateResult
	if err := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni1}}, &outNi); err != nil {
		t.Fatalf("Create network_interface: %v", err)
	}
	idNi := outNi.IDs[0]

	ni2 := iface(d, "wifi2", "4A:C5:93:7A:12:DE", "")
	err2 := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni2}}, nil)
	if !isValidation(err2, devicemanager.ErrRandomizedMAC) {
		t.Errorf("randomized MAC: err = %v, want ValidationError{ErrRandomizedMAC}", err2)
	}

	ni3 := iface(d, "dup", "48:F1:7F:D9:D7:B7", "")
	err3 := callSync(c, opCreate, &crud.CreateArgs{Records: []model.Model{&ni3}}, nil)
	if !isErr(err3, devicemanager.ErrMACAlreadyExists) {
		t.Errorf("duplicate MAC: err = %v, want ErrMACAlreadyExists", err3)
	}

	opUpdate := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Update}
	errUpdate := callSync(c, opUpdate, &crud.UpdateArgs{IDs: []string{idNi}, Fields: []string{"label"}, Record: &devicemanager.NetworkInterface{Label: "wifi intel"}}, nil)
	if errUpdate != nil {
		t.Errorf("update keeping its own MAC: %v", errUpdate)
	}
}

func TestInterface_DeleteDeviceDeletesInterfaces(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	d, _ := deviceInZone(t, c)

	opCreateNi := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni := iface(d, "wifi", "", "")
	if err := callSync(c, opCreateNi, &crud.CreateArgs{Records: []model.Model{&ni}}, nil); err != nil {
		t.Fatalf("Create network_interface: %v", err)
	}

	opDeleteDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Delete}
	if err := callSync(c, opDeleteDev, &crud.DeleteArgs{IDs: []string{d.Id}}, nil); err != nil {
		t.Fatalf("DeleteDevice: %v", err)
	}

	opReadNi := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Read}
	var left devicemanager.NetworkInterfaceList
	err := callSync(c, opReadNi, &crud.ReadArgs{Filter: &devicemanager.NetworkInterface{DeviceId: d.Id}}, &left)
	if err != nil || len(left) != 0 {
		t.Fatalf("interfaces after deleting the device = %+v (err %v), want none", left, err)
	}
}

func TestReaders_FindByIP(t *testing.T) {
	m := setup(t)
	c := crud.LoopbackActingAs(nil, "u1", m)
	d, _ := deviceInZone(t, c)

	opCreateNi := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni := iface(d, "wifi", "", "172.0.0.61")
	if err := callSync(c, opCreateNi, &crud.CreateArgs{Records: []model.Model{&ni}}, nil); err != nil {
		t.Fatalf("Create network_interface: %v", err)
	}
	got, err := m.FindByIP("tenant-A", "172.0.0.61")
	if err != nil || got.Id != d.Id {
		t.Fatalf("FindByIP = %+v (err %v), want device %s", got, err, d.Id)
	}
}

// isValidation reporta si err es un ValidationError que envuelve a want.
func isValidation(err, want error) bool {
	v, ok := err.(*model.FieldError)
	return ok && v.Err == want
}
