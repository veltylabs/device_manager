package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/storage"
	"webtyp.com/time"
)

// Create es la regla de negocio de crear network_interface: valida lo que solo tu dominio sabe y responde si se guarda y si se avisa.
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (p *NetworkInterface) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	var dev Device
	if err := tx.ReadOne(&dev, p.DeviceId); err != nil {
		return crud.RuleAnswer{}, err
	}

	if p.Mac != "" {
		p.Mac = input.CanonicalMAC(p.Mac)
		if input.IsLocallyAdministeredMAC(p.Mac) {
			return crud.RuleAnswer{}, &model.FieldError{Field: "mac", Err: ErrRandomizedMAC}
		}
		var dupCount int
		qDup := crud.Query{Conditions: []storage.Condition{
			storage.Eq(NetworkInterface_.Mac, p.Mac),
			storage.Neq(NetworkInterface_.Id, p.Id),
		}, Limit: 1}
		err := tx.ReadAll(func() model.Model { return &NetworkInterface{} }, qDup, func(row model.Model) {
			dupCount++
		})
		if err != nil {
			return crud.RuleAnswer{}, err
		}
		if dupCount > 0 {
			return crud.RuleAnswer{}, ErrMACAlreadyExists
		}
	}

	var all []NetworkInterface
	qAll := crud.Query{}
	err := tx.ReadAll(func() model.Model { return &NetworkInterface{} }, qAll, func(row model.Model) {
		all = append(all, *row.(*NetworkInterface))
	})
	if err != nil {
		return crud.RuleAnswer{}, err
	}

	var zone Zone
	hasZone := dev.ZoneId != ""
	if hasZone {
		if err := tx.ReadOne(&zone, dev.ZoneId); err != nil {
			return crud.RuleAnswer{}, err
		}
	}

	if p.Ip == "" {
		if !hasZone {
			return crud.RuleAnswer{}, &model.FieldError{Field: "ip", Err: ErrNoZone}
		}
		ip, ok := freeIP(zone, all, p.Id)
		if !ok {
			return crud.RuleAnswer{}, ErrZoneFull
		}
		p.Ip = ip
	} else {
		p.Ip = input.CanonicalIP(p.Ip)
		if hasZone {
			n, okIP := ipv4ToUint32(p.Ip)
			s, e, okZone := zoneRange(zone)
			if !okIP || !okZone || n < s || n > e {
				return crud.RuleAnswer{}, &model.FieldError{Field: "ip", Err: ErrIPOutsideZone}
			}
		}
		for _, o := range all {
			if o.Id != p.Id && o.Ip == p.Ip {
				return crud.RuleAnswer{}, ErrIPAlreadyExists
			}
		}
	}

	p.Id = m.ids.NewID()
	p.UpdatedAt = time.NowNano()
	return crud.ApplyAndPublish(), nil
}
