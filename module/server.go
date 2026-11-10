//go:build !browser

package module

import (
		"webtyp.com/model"
	"webtyp.com/crud"
	"webtyp.com/router"
	"webtyp.com/view"

	devicemanager "github.com/veltylabs/device_manager"
)

// server es lo que el módulo aporta al servidor: cómo se construye.
var server = view.ModuleServer{
	Build: build,
}

// En el servidor el módulo no aporta pantallas.

func build(d appkit.Deps) (model.ModuleNaming, error) {
	m, err := devicemanager.New(d.DB, devicemanager.Deps{IDs: d.IDs, Publisher: d.Events, TenantID: d.TenantID})
	if err != nil {
		return nil, err
	}
	return served{Module: m, tenantID: d.TenantID}, nil
}

// served es el módulo construido: publica sus operaciones y además responde
// "¿qué equipo tiene esta IP?" y "registra este computador" para el tenant de la
// instalación (staffmanager.DeviceReader y staffmanager.DeviceRegistrar), sin que
// staff_manager importe device_manager.
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

// lanInterfaceLabel nombra la interfaz que recibe un computador registrado desde su IP.
const lanInterfaceLabel = "lan"

// RegisterComputer registra un computador activo con una interfaz "lan" en la IP ip
// y devuelve su id. Sin MAC a propósito: el servidor no conoce la MAC de quien se
// conecta, así que esta interfaz sirve para el login por IP y no se envía al router.
//
// Ejemplo: id, err := s.RegisterComputer("Ana (admin)", "192.168.1.50")
func (s served) RegisterComputer(name, ip string) (string, error) {
	dev, err := s.CreateDevice(devicemanager.Device{
		TenantId: s.tenantID, Name: name, Type: devicemanager.DeviceTypeComputer, IsActive: true,
	})
	if err != nil {
		return "", err
	}
	if _, err := s.SaveNetworkInterface(devicemanager.NetworkInterface{
		TenantId: s.tenantID, DeviceId: dev.Id, Label: lanInterfaceLabel, Ip: ip,
	}); err != nil {
		return "", err
	}
	return dev.Id, nil
}
