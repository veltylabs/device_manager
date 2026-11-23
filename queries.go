package devicemanager

import "webtyp.com/fmt"

// lanInterfaceLabel nombra la interfaz que recibe un computador registrado desde su IP.
const lanInterfaceLabel = "lan"

// TenantDevices son los equipos de un tenant: responde «¿qué equipo tiene esta IP?» y «registra este computador»
// con la forma que staff_manager declara (DeviceReader y DeviceRegistrar), sin que staff_manager importe este módulo.
// La app lo arma una vez y se lo pasa a quien lo pide.
//
// Ejemplo: devices := devicemanager.TenantDevices{Devices: dm, TenantID: d.TenantID}
type TenantDevices struct {
	Devices  *Module
	TenantID string
}

// FindByIP devuelve el id del equipo que tiene la IP ip en el tenant.
//
// Ejemplo: id, ok := devices.FindByIP("192.168.1.50")
func (t TenantDevices) FindByIP(ip string) (string, bool) {
	d, err := t.Devices.FindByIP(t.TenantID, ip)
	if err != nil {
		return "", false
	}
	return d.Id, true
}

// RegisterComputer registra un computador activo con una interfaz "lan" en la IP ip y devuelve su id. Sin MAC a
// propósito: el servidor no conoce la MAC de quien se conecta, así que esta interfaz sirve para el ingreso por IP y
// no se envía al router.
//
// Ejemplo: id, err := devices.RegisterComputer("Ana (admin)", "192.168.1.50")
func (t TenantDevices) RegisterComputer(name, ip string) (string, error) {
	dev, err := t.Devices.CreateDevice(Device{TenantId: t.TenantID, Name: name, Type: DeviceTypeComputer, IsActive: true})
	if err != nil {
		return "", err
	}
	if _, err := t.Devices.SaveNetworkInterface(NetworkInterface{
		TenantId: t.TenantID, DeviceId: dev.Id, Label: lanInterfaceLabel, Ip: ip,
	}); err != nil {
		return "", err
	}
	return dev.Id, nil
}

// DeviceOptions entrega los equipos activos del tenant como {id, nombre}: es lo que el mapa de ocupación
// (room_layout.DeviceReader) ofrece al tocar un PC para decir cuál equipo de red es.
//
// Ejemplo: opts, err := dm.DeviceOptions(tenantID)
func (m *Module) DeviceOptions(tenantID string) ([]fmt.KeyValue, error) {
	list, err := m.ListDevices(tenantID, DeviceFilter{ActiveOnly: true})
	if err != nil {
		return nil, err
	}
	out := make([]fmt.KeyValue, 0, len(list))
	for _, d := range list {
		out = append(out, fmt.KeyValue{Key: d.Id, Value: d.Name})
	}
	return out, nil
}
