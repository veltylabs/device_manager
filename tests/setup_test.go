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
	return "test-id-" + fmt.Convert(g.counter).String() // tinywasm/fmt — stdlib strconv is banned, tests included
}

var _ model.IDGenerator = (*mockIDGen)(nil)

type mockPublisher struct{ Events []events.Event }

// events.Publisher.Publish is fire-and-forget: NO error return.
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

// testIfaceLabel is the label of the single interface createWithIP adds.
const testIfaceLabel = "lan"

// createWithIP creates d and gives it one interface with ip — what the tests
// written before interfaces existed meant by "a device with an IP".
func createWithIP(m *devicemanager.Module, d devicemanager.Device, ip string) (devicemanager.Device, error) {
	created, err := m.CreateDevice(d)
	if err != nil {
		return devicemanager.Device{}, err
	}
	_, err = m.SaveNetworkInterface(devicemanager.NetworkInterface{TenantId: d.TenantId, DeviceId: created.Id, Label: testIfaceLabel, Ip: ip})
	return created, err
}
