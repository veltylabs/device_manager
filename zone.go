package devicemanager

import (
	"webtyp.com/events"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

func (m *Module) GetZone(tenantId, id string) (Zone, error) {
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

func (m *Module) ListZones(tenantId string) ([]Zone, error) {
	var z Zone
	rows, err := ReadAllZone(m.db.Query(&z).Where(Zone_.TenantId).Eq(tenantId))
	if err != nil {
		return nil, err
	}
	zones := make([]Zone, len(rows))
	for i, r := range rows {
		zones[i] = *r
	}
	return zones, nil
}

// SaveZone creates the zone when Id is empty, else updates it. Its range must
// be two IPv4 addresses with start <= end that overlap no other zone of the
// tenant.
func (m *Module) SaveZone(z Zone) (Zone, error) {
	creating := z.Id == ""
	z.RangeStart = input.CanonicalIP(z.RangeStart)
	z.RangeEnd = input.CanonicalIP(z.RangeEnd)
	start, end, ok := zoneRange(z)
	if !ok {
		return Zone{}, ValidationError{Err: ErrInvalidZoneRange}
	}
	action := model.ActionUpdate
	if creating {
		z.Id = m.ids.NewID()
		action = model.ActionCreate
	}
	z.UpdatedAt = time.Now()
	if err := z.Validate(action); err != nil {
		return Zone{}, ValidationError{Err: err}
	}
	if !creating {
		if _, err := m.GetZone(z.TenantId, z.Id); err != nil {
			return Zone{}, err
		}
	}
	others, err := m.ListZones(z.TenantId)
	if err != nil {
		return Zone{}, err
	}
	for _, o := range others {
		if o.Id == z.Id {
			continue
		}
		os, oe, ok := zoneRange(o)
		if ok && start <= oe && os <= end {
			return Zone{}, ErrZoneOverlap
		}
	}
	if creating {
		err = m.db.Create(&z)
	} else {
		err = m.db.Update(&z, orm.Eq(Zone_.Id, z.Id), orm.Eq(Zone_.TenantId, z.TenantId))
	}
	if err != nil {
		return Zone{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicZoneSaved, Payload: &z})
	}
	return z, nil
}

// DeleteZone deletes a zone no device uses.
func (m *Module) DeleteZone(tenantId, id string) error {
	z, err := m.GetZone(tenantId, id)
	if err != nil {
		return err
	}
	var d Device
	users, err := ReadAllDevice(m.db.Query(&d).Where(Device_.ZoneId).Eq(id).Where(Device_.TenantId).Eq(tenantId).Limit(1))
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return ErrZoneInUse
	}
	if err := m.db.Delete(&z, orm.Eq(Zone_.Id, z.Id), orm.Eq(Zone_.TenantId, tenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicZoneDeleted, Payload: &z})
	}
	return nil
}

// zoneContaining returns the tenant's zone whose range holds ip, if any.
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
