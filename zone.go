package devicemanager

import (
	"webtyp.com/orm"
)

// GetZone obtiene una zona por su identificador para un tenant dado.
//
// Ejemplo: z, err := dm.GetZone(tenantID, zoneID)
func (m *Module) GetZone(tenantId, id string) (Zone, error) {
	if tenantId != m.tenantID {
		return Zone{}, ErrZoneNotFound
	}
	var z Zone
	qb := m.db.Query(&z).Where(Zone_.Id).Eq(id).Where(Zone_.TenantId).Eq(tenantId)
	if _, err := ReadOneZone(qb, &z); err != nil {
		if orm.IsNotFound(err) {
			return Zone{}, ErrZoneNotFound
		}
		return Zone{}, err
	}
	return z, nil
}

// ListZones lista todas las zonas configuradas para un tenant.
//
// Ejemplo: zonas, err := dm.ListZones(tenantID)
func (m *Module) ListZones(tenantId string, search, after string, limit int) ([]Zone, error) {
	if tenantId != m.tenantID {
		return nil, nil
	}
	var z Zone
	qb := m.db.Query(&z).Where(Zone_.TenantId).Eq(tenantId)
	if search != "" {
		qb = qb.Where(Zone_.Name).Like("%" + search + "%")
	}
	if after != "" {
		qb = qb.Where(Zone_.Id).Gt(after)
	}
	qb = qb.OrderBy(Zone_.Id).Asc().Limit(limit)

	rows, err := ReadAllZone(qb)
	if err != nil {
		return nil, err
	}
	zones := make([]Zone, len(rows))
	for i, r := range rows {
		zones[i] = *r
	}
	return zones, nil
}

// zoneContaining retorna la zona del tenant cuyo rango contiene la ip dada, si existe.
func zoneContaining(zones []Zone, ip string) (Zone, bool) {
	n, ok := ipv4ToUint32(ip)
	if !ok {
		return Zone{}, false
	}
	for _, z := range zones {
		if s, e, ok := zoneRange(z); ok && n >= s && n <= e {
			return z, true
		}
	}
	return Zone{}, false
}
