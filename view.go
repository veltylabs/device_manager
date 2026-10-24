package devicemanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

// Item implementa view.Itemizer para representar un equipo en listas.
//
// Ejemplo: item := d.Item()
func (d *Device) Item() view.Item {
	desc := d.Access
	if d.Location != "" {
		desc += " · " + d.Location
	}
	return view.Item{ID: d.Id, Label: d.Name, Description: desc}
}

// Item implementa view.Itemizer para representar una zona en listas.
//
// Ejemplo: item := z.Item()
func (z *Zone) Item() view.Item {
	return view.Item{ID: z.Id, Label: z.Name, Description: z.RangeStart + " – " + z.RangeEnd}
}

// noMACLabel describe una interfaz que el router no gestiona (sin MAC).
const noMACLabel = "sin MAC"

// Item implementa view.Itemizer para representar una interfaz de red en listas.
//
// Ejemplo: item := ni.Item()
func (ni *NetworkInterface) Item() view.Item {
	mac := ni.Mac
	if mac == "" {
		mac = noMACLabel
	}
	return view.Item{ID: ni.Id, Label: ni.Label, Description: ni.Ip + " · " + mac}
}

const titleDevices = "Equipos"

// NewView construye el Presenter de equipos — el motor independiente de la tecnología
// que cualquier renderizador (crudview u otro) envuelve.
//
// Ejemplo: p := devicemanager.NewView(caller)
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

// NewZoneView construye el Presenter de zonas.
//
// Ejemplo: p := devicemanager.NewZoneView(caller)
func NewZoneView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListZones, Save: OpSaveZone, Delete: OpDeleteZone},
		func() model.ModelSlice { return &ZoneList{} })
	return view.New(b, &Zone{}, view.WithTitle(titleZones))
}

// NewNetworkInterfaceView construye el Presenter de interfaces de red.
//
// Ejemplo: p := devicemanager.NewNetworkInterfaceView(caller)
func NewNetworkInterfaceView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkInterfaces, Save: OpSaveNetworkInterface, Delete: OpDeleteNetworkInterface},
		func() model.ModelSlice { return &NetworkInterfaceList{} })
	return view.New(b, &NetworkInterface{}, view.WithTitle(titleNetworkInterfaces))
}
