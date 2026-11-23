package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

// Al tocar un PC en el mapa de ocupación (room_layout) se elige cuál equipo de red es: DeviceOptions entrega los
// equipos activos del tenant como {id, nombre}; uno dado de baja no se ofrece.
func TestDeviceOptions_ActiveDevicesToChoose(t *testing.T) {
	dm, err := devicemanager.New(orm.New(mem.New()), devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-1"})
	if err != nil {
		t.Fatal(err)
	}
	box2, err := dm.CreateDevice(devicemanager.Device{TenantId: "tenant-1", Name: "PC Box 2", Type: devicemanager.DeviceTypeComputer, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dm.CreateDevice(devicemanager.Device{TenantId: "tenant-1", Name: "PC viejo", Type: devicemanager.DeviceTypeComputer, IsActive: false}); err != nil {
		t.Fatal(err)
	}
	opts, err := dm.DeviceOptions("tenant-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 1 || opts[0].Key != box2.Id || opts[0].Value != "PC Box 2" {
		t.Fatalf("DeviceOptions = %+v, want only the active PC Box 2", opts)
	}
}
