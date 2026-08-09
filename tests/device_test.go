package tests

import (
	"testing"

	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage/mem"
	devicemanager "github.com/veltylabs/device_manager"
)

func TestNew_RequiresIDs(t *testing.T) {
	db := orm.New(mem.New())
	if _, err := devicemanager.New(db, devicemanager.Deps{}); err == nil {
		t.Fatal("expected an error when Deps.IDs is nil")
	}
}

func TestCreateDevice_HappyPath(t *testing.T) {
	m := setup(t)
	d, err := m.CreateDevice(devicemanager.Device{
		TenantId: "tenant-A",
		Name:     "Pc Recepcion",
		Ip:       "192.168.1.10",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if d.Id == "" {
		t.Fatal("expected a generated Id")
	}

	got, err := m.GetDevice("tenant-A", d.Id)
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if got.Name != "Pc Recepcion" || got.Ip != "192.168.1.10" {
		t.Fatalf("unexpected device: %+v", got)
	}
}

func TestCreateDevice_InvalidType(t *testing.T) {
	m := setup(t)
	_, err := m.CreateDevice(devicemanager.Device{
		TenantId: "tenant-A",
		Name:     "Bad",
		Ip:       "192.168.1.11",
		Type:     "not-a-real-type",
		IsActive: true,
	})
	if _, ok := err.(devicemanager.ValidationError); !ok {
		t.Fatalf("expected ValidationError for an invalid type, got %v (%T)", err, err)
	}
}

func TestCreateDevice_DuplicateIPSameTenant(t *testing.T) {
	m := setup(t)
	base := devicemanager.Device{TenantId: "tenant-A", Name: "First", Ip: "192.168.1.12", Type: devicemanager.DeviceTypeServer, IsActive: true}
	if _, err := m.CreateDevice(base); err != nil {
		t.Fatalf("CreateDevice (first): %v", err)
	}
	_, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Second", Ip: "192.168.1.12", Type: devicemanager.DeviceTypeServer, IsActive: true})
	if err != devicemanager.ErrIPAlreadyExists {
		t.Fatalf("expected ErrIPAlreadyExists, got %v", err)
	}
}

func TestCreateDevice_SameIPDifferentTenants_Allowed(t *testing.T) {
	m := setup(t)
	if _, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Device A", Ip: "10.0.0.5", Type: devicemanager.DeviceTypePrinter, IsActive: true}); err != nil {
		t.Fatalf("CreateDevice tenant-A: %v", err)
	}
	if _, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-B", Name: "Device B", Ip: "10.0.0.5", Type: devicemanager.DeviceTypePrinter, IsActive: true}); err != nil {
		t.Fatalf("CreateDevice tenant-B (same IP, different tenant): %v", err)
	}
}

func TestUpdateDevice_NotFound(t *testing.T) {
	m := setup(t)
	_, err := m.UpdateDevice(devicemanager.Device{Id: "does-not-exist", TenantId: "tenant-A", Name: "Device X", Ip: "1.2.3.4", Type: devicemanager.DeviceTypeOther, IsActive: true})
	if err != devicemanager.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeactivateDevice(t *testing.T) {
	m := setup(t)
	d, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Device X", Ip: "1.2.3.5", Type: devicemanager.DeviceTypeOther, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if err := m.DeactivateDevice("tenant-A", d.Id); err != nil {
		t.Fatalf("DeactivateDevice: %v", err)
	}
	got, err := m.GetDevice("tenant-A", d.Id)
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if got.IsActive {
		t.Fatal("expected IsActive=false after DeactivateDevice")
	}
}

func TestDeleteDevice(t *testing.T) {
	m := setup(t)
	d, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Device X", Ip: "1.2.3.6", Type: devicemanager.DeviceTypeOther, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if err := m.DeleteDevice("tenant-A", d.Id); err != nil {
		t.Fatalf("DeleteDevice: %v", err)
	}
	if _, err := m.GetDevice("tenant-A", d.Id); err != devicemanager.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestListDevices_FilterByTypeAndActive(t *testing.T) {
	m := setup(t)
	if _, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Srv1", Ip: "10.1.1.1", Type: devicemanager.DeviceTypeServer, IsActive: true}); err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	pc, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Pc1", Ip: "10.1.1.2", Type: devicemanager.DeviceTypeComputer, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if err := m.DeactivateDevice("tenant-A", pc.Id); err != nil {
		t.Fatalf("DeactivateDevice: %v", err)
	}

	servers, err := m.ListDevices("tenant-A", devicemanager.DeviceFilter{Type: devicemanager.DeviceTypeServer})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(servers) != 1 || servers[0].Type != devicemanager.DeviceTypeServer {
		t.Fatalf("expected 1 server, got %+v", servers)
	}

	active, err := m.ListDevices("tenant-A", devicemanager.DeviceFilter{ActiveOnly: true})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(active) != 1 || active[0].Id != servers[0].Id {
		t.Fatalf("expected only the still-active server, got %+v", active)
	}
}

func TestCreateDevice_PublishesEvent(t *testing.T) {
	db := orm.New(mem.New())
	pub := &mockPublisher{}
	m, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, Publisher: pub})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	d, err := m.CreateDevice(devicemanager.Device{TenantId: "tenant-A", Name: "Device X", Ip: "1.2.3.7", Type: devicemanager.DeviceTypeOther, IsActive: true})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if len(pub.Events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.Events))
	}
	event := pub.Events[0]
	if event.Topic != devicemanager.TopicDeviceCreated {
		t.Fatalf("expected topic %q, got %q", devicemanager.TopicDeviceCreated, event.Topic)
	}
	payload, ok := event.Payload.(*devicemanager.Device)
	if !ok {
		t.Fatalf("expected payload type *Device, got %T", event.Payload)
	}
	if payload.Id != d.Id {
		t.Fatalf("expected payload Id %q, got %q", d.Id, payload.Id)
	}
}
