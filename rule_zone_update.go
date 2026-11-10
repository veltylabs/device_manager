package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/time"
)

// Update es la regla de negocio de actualizar zone: old es la fila guardada y el receptor trae solo los campos enviados (fields).
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (next *Zone) Update(a crud.CurrentUser, tx crud.Tx, old *Zone, fields []string, m *Module) (crud.RuleAnswer, error) {
	z := *old
	for _, f := range fields {
		switch f {
		case "name":
			z.Name = next.Name
		case "range_start":
			z.RangeStart = next.RangeStart
		case "range_end":
			z.RangeEnd = next.RangeEnd
		}
	}
	*next = z // merge

	next.RangeStart = input.CanonicalIP(next.RangeStart)
	next.RangeEnd = input.CanonicalIP(next.RangeEnd)
	start, end, ok := zoneRange(*next)
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
		if o.Id == next.Id {
			continue
		}
		os, oe, ok := zoneRange(o)
		if ok && start <= oe && os <= end {
			return 0, ErrZoneOverlap
		}
	}

	next.UpdatedAt = time.NowNano()
	return crud.ApplyAndPublish, nil
}
