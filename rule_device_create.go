package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/time"
)

// Create es la regla de negocio de crear device: valida lo que solo tu dominio sabe y responde si se guarda y si se avisa.
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (p *Device) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	if !isValidDeviceType(p.Type) {
		return 0, &model.FieldError{Field: "type", Err: fmt.Err("invalid device type: %s", p.Type)}
	}
	if p.Access == "" {
		p.Access = network.AccessLocalName
	}
	if _, err := network.ParseAccess(p.Access); err != nil {
		return 0, err
	}
	if p.ZoneId != "" {
		var z Zone
		if err := tx.ReadOne(&z, p.ZoneId); err != nil {
			return 0, &model.FieldError{Field: "zone_id", Err: err}
		}
	}
	p.UpdatedAt = time.NowNano()
	return crud.ApplyAndPublish, nil
}
