package tests

import (
	"testing"
	"webtyp.com/router"
	"webtyp.com/crud"
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/router/mock"
)

func TestMount_MatchesDefinition(t *testing.T) {
	m := setup(t)
	r := &mock.Router{}
	if err := crud.Mount(r, nil, nil, m); err != nil {
		t.Fatalf("crud.Mount returned error: %v", err)
	}
}

func TestTenant_ReadOtherTenantIsNotFound(t *testing.T) {
	db := orm.New(mem.New())
	mA, _ := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-A"})
	cA := crud.LoopbackActingAs(nil, "u1", mA)

	opCreate := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
	d := &devicemanager.Device{Name: "Device A", Type: devicemanager.DeviceTypeOther, IsActive: true}
	var out crud.CreateResult
	if err := callSync(cA, opCreate, &crud.CreateArgs{Records: []model.Model{d}}, &out); err != nil {
		t.Fatalf("CreateDevice tenant A: %v", err)
	}
	id := out.IDs[0]

	mB, _ := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-B"})

	if _, err := mB.GetDevice("tenant-B", id); !isErr(err, devicemanager.ErrNotFound) {
		t.Fatalf("tenant B read tenant A device: expected ErrNotFound, got %v", err)
	}
}
