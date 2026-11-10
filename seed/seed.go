package seed

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/fmt"
	"webtyp.com/network"
	"webtyp.com/crud"
	"webtyp.com/router"
	"webtyp.com/model"
)

// Direcciones IP y MAC fijas de demostración (administradas globalmente, no aleatorias).
//
// Ejemplo: ip := seed.ReceptionIP
// Prefijos de los errores de Load: el paso que falló, como dato (variables: langc
// traduce las constantes que recibe fmt.Err, y esto es código, no texto).
var (
	errSeedSaveZone             = "seed: SaveZone"
	errSeedCreateDevice         = "seed: CreateDevice"
	errSeedSaveNetworkInterface = "seed: SaveNetworkInterface"
)

const (
	ReceptionIP = "192.168.1.10"
	Box1IP      = "192.168.1.11"
	PrinterIP   = "192.168.1.20"

	ReceptionMAC = "00:1A:2B:00:01:10"
	Box1MAC      = "00:1A:2B:00:01:11"
	PrinterMAC   = "00:1A:2B:00:01:20"

	zoneName       = "Recepción"
	zoneRangeStart = "192.168.1.10"
	zoneRangeEnd   = "192.168.1.29"
	ethernetLabel  = "ethernet"
)

// Data agrupa los registros sembrados para la demostración.
//
// Ejemplo: data := seed.Data{}
type Data struct {
	Zones      []devicemanager.Zone
	Devices    []devicemanager.Device
	Interfaces []devicemanager.NetworkInterface
}

// Load crea la zona, equipos e interfaces de demostración para el tenantID a través de
// los métodos del propio módulo, garantizando la validación completa de cada fila.
//
// Ejemplo: data, err := seed.Load(dm, tenantID)
func Load(m *devicemanager.Module, tenantID string) (Data, error) {
	c := crud.LoopbackActingAs(nil, "importer", m)
	opCreateZone := router.Operation{Module: devicemanager.ModelName, Entity: "zone", Action: model.Create}
	var zone devicemanager.Zone
	zData := devicemanager.Zone{TenantId: tenantID, Name: zoneName,
		RangeStart: zoneRangeStart, RangeEnd: zoneRangeEnd}
	var zErr error
	c.Call(opCreateZone, &crud.CreateArgs{Records: []model.Model{&zData}}, nil, func(err error) { zErr = err })
	err := zErr
	if err != nil {
		return Data{}, fmt.Err(errSeedSaveZone, zoneName, err)
	}
	res := Data{Zones: []devicemanager.Zone{zone}}

	demo := []struct {
		device  devicemanager.Device
		mac, ip string
	}{
		{devicemanager.Device{TenantId: tenantID, Name: "Recepción", Type: devicemanager.DeviceTypeComputer,
			Location: "Recepción", ZoneId: zone.Id, Access: network.AccessInternetName, IsActive: true}, ReceptionMAC, ReceptionIP},
		{devicemanager.Device{TenantId: tenantID, Name: "Box 1", Type: devicemanager.DeviceTypeComputer,
			Location: "Box 1", ZoneId: zone.Id, Access: network.AccessInternetName, IsActive: true}, Box1MAC, Box1IP},
		{devicemanager.Device{TenantId: tenantID, Name: "Impresora", Type: devicemanager.DeviceTypePrinter,
			Location: "Recepción", ZoneId: zone.Id, Access: network.AccessLocalName, IsActive: true}, PrinterMAC, PrinterIP},
	}
	for _, d := range demo {
		createopCreateDev := router.Operation{Module: devicemanager.ModelName, Entity: "device", Action: model.Create}
		var dErr error
		dData := d.device
		c.Call(opCreateDev, &crud.CreateArgs{Records: []model.Model{&dData}}, nil, func(err error) { dErr = err })
		err = dErr
		d := dData
		if err != nil {
			return Data{}, fmt.Err(errSeedCreateDevice, d.device.Name, err)
		}
		opCreateNi := router.Operation{Module: devicemanager.ModelName, Entity: "network_interface", Action: model.Create}
		var niErr error
		niData := devicemanager.NetworkInterface{TenantId: tenantID,
			DeviceId: d.Id, Label: ethernetLabel, Mac: d.mac, Ip: d.ip}
		c.Call(opCreateNi, &crud.CreateArgs{Records: []model.Model{&niData}}, nil, func(err error) { niErr = err })
		err = niErr
		ni := niData
		if err != nil {
			return Data{}, fmt.Err(errSeedSaveNetworkInterface, d.device.Name, d.ip, err)
		}
		res.Devices = append(res.Devices, d)
		res.Interfaces = append(res.Interfaces, ni)
	}
	return res, nil
}
