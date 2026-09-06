package devicemanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

// Item implements view.Itemizer if needed or defined.
func (d *Device) Item() view.Item {
	return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
}

const titleDevices = "Equipos"

// NewView builds the device Presenter — the tech-agnostic engine a renderer
// (crudview, or any other) wraps. This module builds it (view + model + router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListDevices, Save: OpUpsertDevice, Delete: OpDeleteDevice},
		func() model.ModelSlice { return &DeviceList{} })
	return view.New(b, &Device{}, view.WithTitle(titleDevices))
}
