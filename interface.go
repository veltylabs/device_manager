package devicemanager

import (
	"webtyp.com/orm"
)

// GetNetworkInterface obtiene una interfaz de red por su identificador para un tenant dado.
//
// Ejemplo: iface, err := dm.GetNetworkInterface(tenantID, ifaceID)
func (m *Module) GetNetworkInterface(tenantId, id string) (NetworkInterface, error) {
	if tenantId != m.tenantID {
		return NetworkInterface{}, nil
	}
	var ni NetworkInterface
	qb := m.db.Query(&ni).Where(NetworkInterface_.Id).Eq(id).Where(NetworkInterface_.TenantId).Eq(tenantId)
	if _, err := ReadOneNetworkInterface(qb, &ni); err != nil {
		if orm.IsNotFound(err) {
			return NetworkInterface{}, ErrInterfaceNotFound
		}
		return NetworkInterface{}, err
	}
	return ni, nil
}

// ListNetworkInterfaces lista las interfaces de red del tenant; si se especifica deviceId,
// acota los resultados a las interfaces de dicho equipo.
//
// Ejemplo: ifaces, err := dm.ListNetworkInterfaces(tenantID, deviceID)
func (m *Module) ListNetworkInterfaces(tenantId, deviceId, search, after string, limit int) ([]NetworkInterface, error) {
	if tenantId != m.tenantID {
		return nil, nil
	}
	var ni NetworkInterface
	qb := m.db.Query(&ni).Where(NetworkInterface_.TenantId).Eq(tenantId)
	if deviceId != "" {
		qb = qb.Where(NetworkInterface_.DeviceId).Eq(deviceId)
	}
	if search != "" {
		qb = qb.Where(NetworkInterface_.Label).Like("%" + search + "%")
	}
	if after != "" {
		qb = qb.Where(NetworkInterface_.Id).Gt(after)
	}
	qb = qb.OrderBy(NetworkInterface_.Id).Asc().Limit(limit)

	rows, err := ReadAllNetworkInterface(qb)
	if err != nil {
		return nil, err
	}
	out := make([]NetworkInterface, len(rows))
	for i, r := range rows {
		out[i] = *r
	}
	return out, nil
}





// freeIP retorna la dirección más baja de la zona que no esté asignada a ninguna otra interfaz distinta de self.
func freeIP(z Zone, all []NetworkInterface, self string) (string, bool) {
	s, e, ok := zoneRange(z)
	if !ok {
		return "", false
	}
	for n := uint64(s); n <= uint64(e); n++ {
		ip := uint32ToIPv4(uint32(n))
		taken := false
		for _, o := range all {
			if o.Id != self && o.Ip == ip {
				taken = true
				break
			}
		}
		if !taken {
			return ip, true
		}
	}
	return "", false
}
