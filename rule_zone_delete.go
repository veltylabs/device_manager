package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
)

func (p *Zone) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	count := 0
	q := crud.Query{Conditions: []storage.Condition{storage.Eq(Device_.ZoneId, p.Id)}, Limit: 1}
	err := tx.ReadAll(func() model.Model { return &Device{} }, q, func(r model.Model) {
		count++
	})
	if err != nil {
		return crud.RuleAnswer{}, err
	}
	if count > 0 {
		return crud.RuleAnswer{}, ErrZoneInUse
	}
	return crud.ApplyAndPublish(), nil
}
