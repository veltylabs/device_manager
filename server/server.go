//go:build !wasm

// Package server es device_manager tal como lo enchufa una app: se migra y se
// construye con lo que entrega appkit.
package server

import (
	"webtyp.com/appkit"
	"webtyp.com/ddl"
	"webtyp.com/model"

	devicemanager "github.com/veltylabs/device_manager"
	"github.com/veltylabs/device_manager/migrate"
)

// Module es device_manager para appkit: va en App.Modules de la app.
//
// Ejemplo: Modules: []appkit.Module{devices.Module, staff.Module}
var Module = appkit.Module{
	Name: "device_manager",
	Migrate: func(conn ddl.Execer, c ddl.Compiler) error {
		if err := migrate.Migrate(conn, c); err != nil {
			return err
		}
		return migrate.LegacyDeviceIP(conn, c)
	},
	Build: build,
}

func build(d appkit.Deps) (model.ModuleNaming, error) {
	m, err := devicemanager.New(d.DB, devicemanager.Deps{IDs: d.IDs, Publisher: d.Events, TenantID: d.Tenant})
	if err != nil {
		return nil, err
	}
	return served{Module: m, tenantID: d.Tenant}, nil
}

// served es el módulo construido: publica sus operaciones y además responde
// "¿qué equipo tiene esta IP?" (staffmanager.DeviceReader) para el tenant de la
// instalación, sin que staff_manager importe device_manager.
type served struct {
	*devicemanager.Module
	tenantID string
}

// FindByIP devuelve el id del equipo que tiene la IP ip.
//
// Ejemplo: id, ok := s.FindByIP("192.168.1.50")
func (s served) FindByIP(ip string) (string, bool) {
	return devicemanager.IPLocator{Devices: s.Module, TenantID: s.tenantID}.FindByIP(ip)
}
