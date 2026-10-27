package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

type mockIDGen struct{ counter int }

func (g *mockIDGen) NewID() string {
	g.counter++
	return "test-id-" + fmt.Convert(g.counter).String() // tinywasm/fmt — la librería estándar strconv está prohibida, incluso en tests
}

var _ model.IDGenerator = (*mockIDGen)(nil)

type mockPublisher struct{ Events []events.Event }

// events.Publisher.Publish dispara y olvida: NO retorna error.
func (p *mockPublisher) Publish(e events.Event) {
	p.Events = append(p.Events, e)
}

var _ events.Publisher = (*mockPublisher)(nil)

func setup(t *testing.T) *devicemanager.Module {
	t.Helper()
	db := orm.New(mem.New())
	m, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: "tenant-A"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}

// testIfaceLabel es la etiqueta de la única interfaz que agrega createWithIP.
const testIfaceLabel = "lan"

// createWithIP crea d y le asigna una interfaz con ip — a lo que los tests
// escritos antes de que existieran las interfaces se referían como "un equipo con IP".
func createWithIP(m *devicemanager.Module, d devicemanager.Device, ip string) (devicemanager.Device, error) {
	created, err := m.CreateDevice(d)
	if err != nil {
		return devicemanager.Device{}, err
	}
	_, err = m.SaveNetworkInterface(devicemanager.NetworkInterface{TenantId: d.TenantId, DeviceId: created.Id, Label: testIfaceLabel, Ip: ip})
	return created, err
}
