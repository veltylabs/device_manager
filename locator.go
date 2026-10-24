package devicemanager

// IPLocator responde "¿qué equipo tiene esta IP?" para un tenant dado. Cumple
// exactamente la forma requerida por el puerto DeviceReader de staff_manager,
// satisfaciéndolo estructuralmente sin acoplamiento directo entre módulos.
//
// Ejemplo: loc := devicemanager.IPLocator{Devices: dm, TenantID: tenantID}
type IPLocator struct {
	Devices  *Module
	TenantID string
}

// FindByIP busca el ID del equipo asociado a una IP en el tenant configurado.
//
// Ejemplo: deviceID, ok := loc.FindByIP("192.168.1.50")
func (l IPLocator) FindByIP(ip string) (string, bool) {
	d, err := l.Devices.FindByIP(l.TenantID, ip)
	if err != nil {
		return "", false
	}
	return d.Id, true
}
