package tests

import (
	"testing"
	"webtyp.com/crud"
	"webtyp.com/router"

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
func createWithIP(c crud.Caller, d devicemanager.Device, ip string) (devicemanager.Device, error) {
	opCreateDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
	var created devicemanager.Device
	if err := c.Call(opCreateDev, &d, &created); err != nil {
		return devicemanager.Device{}, err
	}

	opCreateIface := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
	ni := devicemanager.NetworkInterface{DeviceId: created.Id, Label: testIfaceLabel, Ip: ip}
	if err := c.Call(opCreateIface, &ni, nil); err != nil {
		return devicemanager.Device{}, err
	}

	return created, nil
}
	_, err = m.SaveNetworkInterface(devicemanager.NetworkInterface{TenantId: d.TenantId, DeviceId: created.Id, Label: testIfaceLabel, Ip: ip})
	return created, err
}

func callSync(c router.Caller, op router.Operation, args model.Encodable, into model.Decodable) error {
	var callErr error
	c.Call(op, args, into, func(err error) {
		callErr = err
	})
	return callErr
}
