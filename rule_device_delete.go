package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
)

// Delete es la regla de negocio de borrar device: sus interfaces de red se borran primero, en la misma
// transacción, porque network_interface.device_id referencia a device.id. El receptor es la fila guardada.
func (p *Device) Delete(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	for {
		var ids []string
		err := tx.ReadAll(func() model.Model { return &NetworkInterface{} },
			crud.Query{Conditions: []storage.Condition{storage.Eq(NetworkInterface_.DeviceId, p.Id)}, Limit: crud.MaxPage},
			func(r model.Model) { ids = append(ids, r.(*NetworkInterface).Id) })
		if err != nil {
			return 0, err
		}
		if len(ids) == 0 {
			return crud.ApplyAndPublish, nil
		}
		for _, id := range ids {
			if err := tx.Delete(&NetworkInterface{}, id); err != nil {
				return 0, err
			}
		}
	}
}
