//go:build wasm

package main

import (
	devicemanager "github.com/veltylabs/device_manager"
	"github.com/veltylabs/device_manager/seed"
	"github.com/veltylabs/device_manager/ui"
	. "webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/layout/platformd"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

// demoTenantID is the only tenant of this in-browser demo.
const demoTenantID = "demo"

// demoUser is the fixed identity the demo shell shows: the demo has no login.
type demoUser struct{}

func (demoUser) UserName() string    { return "Demo" }
func (demoUser) UserAvatar() string  { return "" }
func (demoUser) UserRoles() []string { return []string{"Administrador"} }

func main() {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	db := orm.New(mem.New())
	broker := &mock.Broker{}

	dm, err := devicemanager.New(db, devicemanager.Deps{IDs: ids, Publisher: broker, TenantID: demoTenantID})
	if err != nil {
		panic(err)
	}

	if _, err := seed.Load(dm, demoTenantID); err != nil {
		panic(err)
	}

	caller := loopback.WithTenant(demoTenantID, dm)
	v, err := ui.Browser(caller, ids, demoTenantID)
	if err != nil {
		panic(err)
	}

	p := &platformd.Platform{
		AppName:   ui.Label + " — demo",
		User:      demoUser{},
		Modules:   []platformd.UIModule{v},
		DefaultID: ui.ID,
	}
	Append("body", p)
	select {}
}
