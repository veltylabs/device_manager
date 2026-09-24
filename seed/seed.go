package seed

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/fmt"
)

const (
	ReceptionIP = "192.168.1.10"
	Box1IP      = "192.168.1.11"
	PrinterIP   = "192.168.1.20"
)

type Data struct {
	Devices []devicemanager.Device
}

// Load creates the demo devices for tenantID through m.CreateDevice.
func Load(m *devicemanager.Module, tenantID string) (Data, error) {
	devices := []devicemanager.Device{
		{
			TenantId: tenantID,
			Name:     "Recepción",
			Ip:       ReceptionIP,
			Type:     devicemanager.DeviceTypeComputer,
			Location: "Recepción",
			IsActive: true,
		},
		{
			TenantId: tenantID,
			Name:     "Box 1",
			Ip:       Box1IP,
			Type:     devicemanager.DeviceTypeComputer,
			Location: "Box 1",
			IsActive: true,
		},
		{
			TenantId: tenantID,
			Name:     "Impresora",
			Ip:       PrinterIP,
			Type:     devicemanager.DeviceTypePrinter,
			Location: "Recepción",
			IsActive: true,
		},
	}

	res := Data{
		Devices: make([]devicemanager.Device, 0, len(devices)),
	}

	for _, d := range devices {
		created, err := m.CreateDevice(d)
		if err != nil {
			return Data{}, fmt.Err("seed Load CreateDevice failed for %s (%s): %w", d.Name, d.Ip, err)
		}
		res.Devices = append(res.Devices, created)
	}

	return res, nil
}
