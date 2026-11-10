package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/time"
	"webtyp.com/storage"
)

// Update es la regla de negocio de actualizar network_interface: old es la fila guardada y el receptor trae solo los campos enviados (fields).
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (next *NetworkInterface) Update(a crud.CurrentUser, tx crud.Tx, old *NetworkInterface, fields []string, m *Module) (crud.RuleAnswer, error) {
	ni := *old
	for _, f := range fields {
		switch f {
		case "device_id":
			ni.DeviceId = next.DeviceId
		case "label":
			ni.Label = next.Label
		case "mac":
			ni.Mac = next.Mac
		case "ip":
			ni.Ip = next.Ip
		}
	}
	*next = ni // merge

	var dev Device
	if err := tx.ReadOne(&dev, next.DeviceId); err != nil {
		return 0, err
	}

	if next.Mac != "" {
		next.Mac = input.CanonicalMAC(next.Mac)
		if input.IsLocallyAdministeredMAC(next.Mac) {
			return 0, &model.FieldError{Field: "mac", Err: ErrRandomizedMAC}
		}
		var dupCount int
		qDup := crud.Query{Conditions: []storage.Condition{
			storage.Eq(NetworkInterface_.Mac, next.Mac),
			storage.Eq(NetworkInterface_.TenantId, next.TenantId),
			storage.Neq(NetworkInterface_.Id, next.Id),
		}, Limit: 1}
		err := tx.ReadAll(func() model.Model { return &NetworkInterface{} }, qDup, func(row model.Model) {
			dupCount++
		})
		if err != nil {
			return 0, err
		}
		if dupCount > 0 {
			return 0, ErrMACAlreadyExists
		}
	}

	var all []NetworkInterface
	qAll := crud.Query{Conditions: []storage.Condition{storage.Eq(NetworkInterface_.TenantId, next.TenantId)}}
	err := tx.ReadAll(func() model.Model { return &NetworkInterface{} }, qAll, func(row model.Model) {
		all = append(all, *row.(*NetworkInterface))
	})
	if err != nil {
		return 0, err
	}

	var zone Zone
	hasZone := dev.ZoneId != ""
	if hasZone {
		if err := tx.ReadOne(&zone, dev.ZoneId); err != nil {
			return 0, err
		}
	}

	if next.Ip == "" {
		if !hasZone {
			return 0, &model.FieldError{Field: "ip", Err: ErrNoZone}
		}
		ip, ok := freeIP(zone, all, next.Id)
		if !ok {
			return 0, ErrZoneFull
		}
		next.Ip = ip
	} else {
		next.Ip = input.CanonicalIP(next.Ip)
		if hasZone {
			n, okIP := ipv4ToUint32(next.Ip)
			s, e, okZone := zoneRange(zone)
			if !okIP || !okZone || n < s || n > e {
				return 0, &model.FieldError{Field: "ip", Err: ErrIPOutsideZone}
			}
		}
		for _, o := range all {
			if o.Id != next.Id && o.Ip == next.Ip {
				return 0, ErrIPAlreadyExists
			}
		}
	}

	next.UpdatedAt = time.NowNano()
	return crud.ApplyAndPublish, nil
}
