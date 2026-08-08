package devicemanager

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/view"
)

// Item implements view.Itemizer if needed or defined.
func (d *Device) Item() view.Item {
	return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
}

// NewView builds the device Presenter — the tech-agnostic engine a renderer (tinywasm/layout's
// crudview, or any other) wraps. This module builds it (only view+model+router); the app decides
// which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	var byID []*Device
	record := &Device{}

	return view.New(
		caller,
		record,
		OpListDevices,
		func() model.FielderSlice { return &DeviceList{} },
		func(list model.FielderSlice) []view.Item {
			l := list.(*DeviceList)
			items := make([]view.Item, l.Len())
			byID = make([]*Device, l.Len())
			for i := 0; i < l.Len(); i++ {
				it := l.At(i).(*Device)
				byID[i] = it
				items[i] = view.Item{ID: it.Id, Label: it.Name, Description: it.Ip}
			}
			return items
		},
		view.WithTitle("Equipos"),
		view.WithSaveOp(OpUpsertDevice),
		view.WithDeleteOp(OpDeleteDevice),
		view.WithFill(func(id string) model.Model {
			if id == "" {
				return nil
			}
			for _, it := range byID {
				if it != nil && it.Id == id {
					return it
				}
			}
			return nil
		}),
	)
}
