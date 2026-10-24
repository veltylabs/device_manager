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

// demoTenantID es el único tenant de esta demostración en el navegador.
const demoTenantID = "demo"

// demoUser es la identidad fija que muestra el chasis de demo: la demo no requiere inicio de sesión.
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
	zones, err := ui.ZonesBrowser(caller, ids, demoTenantID)
	if err != nil {
		panic(err)
	}
	ifaces, err := ui.InterfacesBrowser(caller, ids, demoTenantID)
	if err != nil {
		panic(err)
	}

	p := &platformd.Platform{
		AppName:   ui.Label + " — demo",
		User:      demoUser{},
		Modules:   []platformd.UIModule{v, zones, ifaces},
		DefaultID: ui.ID,
	}
	Append("body", p)
	select {}
}
