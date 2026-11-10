package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/storage/mem"
)

func callAs(user string, mod *devicemanager.Module, entity string, action model.Action, args model.Encodable, into model.Decodable) error {
	var callErr error
	crud.LoopbackActingAs(nil, user, mod).Call(
		router.Operation{Module: devicemanager.ModelName, Entity: entity, Action: action},
		args, into, func(err error) { callErr = err })
	return callErr
}

func TestMount_MatchesDefinition(t *testing.T) {
	m := setup(t)
	if err := crud.Mount(nil, nil, nil, m); err != nil {
		t.Fatalf("what MountOperations mounts must match Definition(), got %v", err)
	}
}

func TestTenant_ReadOtherTenantIsNotFound(t *testing.T) {
	db := orm.New(mem.New())
	modA, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-A"})
	if err != nil {
		t.Fatal(err)
	}
	modB, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-B"})
	if err != nil {
		t.Fatal(err)
	}
	created, err := modA.CreateDevice(devicemanager.Device{Name: "PC", Type: devicemanager.DeviceTypeComputer, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	var list devicemanager.DeviceList
	if err := callAs("u1", modA, "device", model.Read, &crud.ReadArgs{ID: created.Id}, &list); err != nil || len(list) != 1 {
		t.Fatalf("tenant A must read its own device, got %d rows err %v", len(list), err)
	}
	err = callAs("u1", modB, "device", model.Read, &crud.ReadArgs{ID: created.Id}, &list)
	if err == nil || model.KindOf(err) != model.ErrorNotFound {
		t.Fatalf("tenant B reading a device of tenant A must be ErrorNotFound, got %v", err)
	}
}

func TestDevice_CreateRejectsClientTenant(t *testing.T) {
	m := setup(t)
	err := callAs("u1", m, "device", model.Create,
		&crud.CreateArgs{Records: []model.Model{&devicemanager.Device{TenantId: "tenant-B", Name: "PC", Type: devicemanager.DeviceTypeComputer, IsActive: true}}}, nil)
	if err == nil {
		t.Fatal("a tenant_id sent by the client must be refused (ServerOnly)")
	}
}

func TestDevice_DeleteRemovesItsInterfaces(t *testing.T) {
	m := setup(t)
	device, err := createWithIP(m, devicemanager.Device{Name: "PC", Type: devicemanager.DeviceTypeComputer, IsActive: true}, "192.168.1.30")
	if err != nil {
		t.Fatalf("createWithIP: %v", err)
	}
	if err := callAs("u1", m, "device", model.Delete, &crud.DeleteArgs{IDs: []string{device.Id}}, nil); err != nil {
		t.Fatalf("delete device: %v", err)
	}
	rows, err := m.ListNetworkInterfaces("tenant-A", device.Id, "", "", 0)
	if err != nil || len(rows) != 0 {
		t.Fatalf("the interfaces of a deleted device must go with it, got %d err %v", len(rows), err)
	}
}
