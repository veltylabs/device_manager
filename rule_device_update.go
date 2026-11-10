package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/time"
)

// Update es la regla de negocio de actualizar device: old es la fila guardada y el receptor trae solo los campos enviados (fields).
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (next *Device) Update(a crud.CurrentUser, tx crud.Tx, old *Device, fields []string, m *Module) (crud.RuleAnswer, error) {
	d := *old
	for _, f := range fields {
		switch f {
		case "name":
			d.Name = next.Name
		case "type":
			d.Type = next.Type
		case "location":
			d.Location = next.Location
		case "zone_id":
			d.ZoneId = next.ZoneId
		case "access":
			d.Access = next.Access
		case "is_active":
			d.IsActive = next.IsActive
		}
	}
	*next = d // merge para comprobaciones y guardado por crud

	if !isValidDeviceType(next.Type) {
		return 0, &model.FieldError{Field: "type", Err: fmt.Err("invalid device type: %s", next.Type)}
	}
	if _, err := network.ParseAccess(next.Access); err != nil {
		return 0, err
	}
	if next.ZoneId != "" {
		var z Zone
		if err := tx.ReadOne(&z, next.ZoneId); err != nil {
			return 0, &model.FieldError{Field: "zone_id", Err: err}
		}
	}
	next.UpdatedAt = time.NowNano()

	// DeactivateDevice logic (is_active = false)
	// No other manual hook is needed, crud will emit event with changes.
	return crud.ApplyAndPublish, nil
}
