package devicemanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

// Item implements view.Itemizer if needed or defined.
func (d *Device) Item() view.Item {
	desc := d.Access
	if d.Location != "" {
		desc += " · " + d.Location
	}
	return view.Item{ID: d.Id, Label: d.Name, Description: desc}
}

func (z *Zone) Item() view.Item {
	return view.Item{ID: z.Id, Label: z.Name, Description: z.RangeStart + " – " + z.RangeEnd}
}

// noMACLabel describes an interface the router does not manage (no MAC).
const noMACLabel = "sin MAC"

func (ni *NetworkInterface) Item() view.Item {
	mac := ni.Mac
	if mac == "" {
		mac = noMACLabel
	}
	return view.Item{ID: ni.Id, Label: ni.Label, Description: ni.Ip + " · " + mac}
}

const titleDevices = "Equipos"

// NewView builds the device Presenter — the tech-agnostic engine a renderer
// (crudview, or any other) wraps. This module builds it (view + model + router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListDevices, Save: OpUpsertDevice, Delete: OpDeleteDevice},
		func() model.ModelSlice { return &DeviceList{} })
	return view.New(b, &Device{}, view.WithTitle(titleDevices))
}

const (
	titleZones             = "Zonas"
	titleNetworkInterfaces = "Interfaces de red"
)

// NewZoneView builds the zones Presenter.
func NewZoneView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListZones, Save: OpSaveZone, Delete: OpDeleteZone},
		func() model.ModelSlice { return &ZoneList{} })
	return view.New(b, &Zone{}, view.WithTitle(titleZones))
}

// NewNetworkInterfaceView builds the network interfaces Presenter.
func NewNetworkInterfaceView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkInterfaces, Save: OpSaveNetworkInterface, Delete: OpDeleteNetworkInterface},
		func() model.ModelSlice { return &NetworkInterfaceList{} })
	return view.New(b, &NetworkInterface{}, view.WithTitle(titleNetworkInterfaces))
}
