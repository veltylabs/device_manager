package seed

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/fmt"
	"webtyp.com/network"
)

const (
	ReceptionIP = "192.168.1.10"
	Box1IP      = "192.168.1.11"
	PrinterIP   = "192.168.1.20"

	// Fixed, globally administered (not randomized) demo MACs.
	ReceptionMAC = "00:1A:2B:00:01:10"
	Box1MAC      = "00:1A:2B:00:01:11"
	PrinterMAC   = "00:1A:2B:00:01:20"

	zoneName       = "Recepción"
	zoneRangeStart = "192.168.1.10"
	zoneRangeEnd   = "192.168.1.29"
	ethernetLabel  = "ethernet"
)

type Data struct {
	Zones      []devicemanager.Zone
	Devices    []devicemanager.Device
	Interfaces []devicemanager.NetworkInterface
}

// Load creates the demo zone, devices and interfaces for tenantID through the
// module's own methods, so every row is validated.
func Load(m *devicemanager.Module, tenantID string) (Data, error) {
	zone, err := m.SaveZone(devicemanager.Zone{TenantId: tenantID, Name: zoneName,
		RangeStart: zoneRangeStart, RangeEnd: zoneRangeEnd})
	if err != nil {
		return Data{}, fmt.Err("seed: SaveZone", zoneName, err)
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
		created, err := m.CreateDevice(d.device)
		if err != nil {
			return Data{}, fmt.Err("seed: CreateDevice", d.device.Name, err)
		}
		ni, err := m.SaveNetworkInterface(devicemanager.NetworkInterface{TenantId: tenantID,
			DeviceId: created.Id, Label: ethernetLabel, Mac: d.mac, Ip: d.ip})
		if err != nil {
			return Data{}, fmt.Err("seed: SaveNetworkInterface", d.device.Name, d.ip, err)
		}
		res.Devices = append(res.Devices, created)
		res.Interfaces = append(res.Interfaces, ni)
	}
	return res, nil
}
