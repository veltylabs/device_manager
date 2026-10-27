package devicemanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titleDevices = "Equipos"

// NewView construye el Presenter de equipos — el motor independiente de la tecnología
// que cualquier renderizador (crudview u otro) envuelve.
//
// Ejemplo: p := devicemanager.NewView(caller, view.WithDescription("Ayuda..."))
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, List: OpListDevices, Save: OpUpsertDevice, Delete: OpDeleteDevice},
		func() model.ModelSlice { return &DeviceList{} })
	allOpts := append([]view.Option{view.WithTitle(titleDevices)}, opts...)
	return view.New(b, &Device{}, allOpts...)
}

const (
	titleZones             = "Zonas"
	titleNetworkInterfaces = "Interfaces de red"
)

// NewZoneView construye el Presenter de zonas.
//
// Ejemplo: p := devicemanager.NewZoneView(caller, view.WithDescription("Ayuda..."))
func NewZoneView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, List: OpListZones, Save: OpSaveZone, Delete: OpDeleteZone},
		func() model.ModelSlice { return &ZoneList{} })
	allOpts := append([]view.Option{view.WithTitle(titleZones)}, opts...)
	return view.New(b, &Zone{}, allOpts...)
}

// NewNetworkInterfaceView construye el Presenter de interfaces de red.
//
// Ejemplo: p := devicemanager.NewNetworkInterfaceView(caller, view.WithDescription("Ayuda..."))
func NewNetworkInterfaceView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, List: OpListNetworkInterfaces, Save: OpSaveNetworkInterface, Delete: OpDeleteNetworkInterface},
		func() model.ModelSlice { return &NetworkInterfaceList{} })
	allOpts := append([]view.Option{view.WithTitle(titleNetworkInterfaces)}, opts...)
	return view.New(b, &NetworkInterface{}, allOpts...)
}
