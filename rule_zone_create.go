package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/time"
)

// Create es la regla de negocio de crear zone: valida lo que solo tu dominio sabe y responde si se guarda y si se avisa.
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (p *Zone) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	p.RangeStart = input.CanonicalIP(p.RangeStart)
	p.RangeEnd = input.CanonicalIP(p.RangeEnd)
	start, end, ok := zoneRange(*p)
	if !ok {
		return 0, &model.FieldError{Field: "range_start", Err: ErrInvalidZoneRange}
	}

	var others []Zone
	q := crud.Query{}
	err := tx.ReadAll(func() model.Model { return &Zone{} }, q, func(row model.Model) {
		others = append(others, *row.(*Zone))
	})
	if err != nil {
		return 0, err
	}

	for _, o := range others {
		if o.Id == p.Id {
			continue
		}
		os, oe, ok := zoneRange(o)
		if ok && start <= oe && os <= end {
			return 0, ErrZoneOverlap
		}
	}

	p.Id = m.ids.NewID()
	p.UpdatedAt = time.NowNano()
	return crud.ApplyAndPublish, nil
}
