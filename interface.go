package devicemanager

import (
	"webtyp.com/events"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
)

func (m *Module) GetNetworkInterface(tenantId, id string) (NetworkInterface, error) {
	var ni NetworkInterface
	qb := m.db.Query(&ni).Where(NetworkInterface_.Id).Eq(id).Where(NetworkInterface_.TenantId).Eq(tenantId)
	if _, err := ReadOneNetworkInterface(qb, &ni); err != nil {
		if err == orm.ErrNotFound {
			return NetworkInterface{}, ErrInterfaceNotFound
		}
		return NetworkInterface{}, err
	}
	return ni, nil
}

// ListNetworkInterfaces lists the tenant's interfaces; a non-empty deviceId
// narrows them to that device.
func (m *Module) ListNetworkInterfaces(tenantId, deviceId string) ([]NetworkInterface, error) {
	var ni NetworkInterface
	qb := m.db.Query(&ni).Where(NetworkInterface_.TenantId).Eq(tenantId)
	if deviceId != "" {
		qb = qb.Where(NetworkInterface_.DeviceId).Eq(deviceId)
	}
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

// SaveNetworkInterface creates the interface when Id is empty, else updates
// it. See docs/ARCHITECTURE.md for the MAC and IP rules it enforces.
func (m *Module) SaveNetworkInterface(ni NetworkInterface) (NetworkInterface, error) {
	creating := ni.Id == ""
	dev, err := m.GetDevice(ni.TenantId, ni.DeviceId)
	if err != nil {
		return NetworkInterface{}, err
	}
	if !creating {
		if _, err := m.GetNetworkInterface(ni.TenantId, ni.Id); err != nil {
			return NetworkInterface{}, err
		}
	}

	if ni.Mac != "" {
		ni.Mac = input.CanonicalMAC(ni.Mac)
		if input.IsLocallyAdministeredMAC(ni.Mac) {
			return NetworkInterface{}, ValidationError{Err: ErrRandomizedMAC}
		}
		var other NetworkInterface
		dup, err := ReadAllNetworkInterface(m.db.Query(&other).
			Where(NetworkInterface_.Mac).Eq(ni.Mac).
			Where(NetworkInterface_.TenantId).Eq(ni.TenantId).
			Where(NetworkInterface_.Id).Neq(ni.Id).Limit(1))
		if err != nil {
			return NetworkInterface{}, err
		}
		if len(dup) > 0 {
			return NetworkInterface{}, ErrMACAlreadyExists
		}
	}

	all, err := m.ListNetworkInterfaces(ni.TenantId, "")
	if err != nil {
		return NetworkInterface{}, err
	}
	var zone Zone
	hasZone := dev.ZoneId != ""
	if hasZone {
		if zone, err = m.GetZone(dev.TenantId, dev.ZoneId); err != nil {
			return NetworkInterface{}, err
		}
	}

	if ni.Ip == "" {
		if !hasZone {
			return NetworkInterface{}, ValidationError{Err: ErrNoZone}
		}
		ip, ok := freeIP(zone, all, ni.Id)
		if !ok {
			return NetworkInterface{}, ErrZoneFull
		}
		ni.Ip = ip
	} else {
		ni.Ip = input.CanonicalIP(ni.Ip)
		if hasZone {
			n, okIP := ipv4ToUint32(ni.Ip)
			s, e, okZone := zoneRange(zone)
			if !okIP || !okZone || n < s || n > e {
				return NetworkInterface{}, ValidationError{Err: ErrIPOutsideZone}
			}
		}
		for _, o := range all {
			if o.Id != ni.Id && o.Ip == ni.Ip {
				return NetworkInterface{}, ErrIPAlreadyExists
			}
		}
	}

	action := model.ActionUpdate
	if creating {
		ni.Id = m.ids.NewID()
		action = model.ActionCreate
	}
	ni.UpdatedAt = time.Now()
	if err := ni.Validate(action); err != nil {
		return NetworkInterface{}, ValidationError{Err: err}
	}
	if creating {
		err = m.db.Create(&ni)
	} else {
		err = m.db.Update(&ni, orm.Eq(NetworkInterface_.Id, ni.Id), orm.Eq(NetworkInterface_.TenantId, ni.TenantId))
	}
	if err != nil {
		return NetworkInterface{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicNetworkInterfaceSaved, Payload: &ni})
	}
	return ni, nil
}

func (m *Module) DeleteNetworkInterface(tenantId, id string) error {
	ni, err := m.GetNetworkInterface(tenantId, id)
	if err != nil {
		return err
	}
	if err := m.db.Delete(&ni, orm.Eq(NetworkInterface_.Id, ni.Id), orm.Eq(NetworkInterface_.TenantId, tenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicNetworkInterfaceDeleted, Payload: &ni})
	}
	return nil
}

// freeIP returns the lowest address of the zone not held by any interface
// other than self.
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
