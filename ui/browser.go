package ui

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

// Campos select cuyas opciones provienen de otra lista de este módulo.
const (
	fieldZoneID   = "zone_id"
	fieldDeviceID = "device_id"
)

// Textos de ayuda que el chasis muestra en el botón «?» de cada pantalla. Las
// claves están en inglés; la traducción vive en lang.json.
const (
	descDevices    lang.Text = "Each computer, printer or device of the clinic, with the zone where it is. Its network interfaces are added in Network interfaces."
	descZones      lang.Text = "Zones group the devices by place (for example, reception or box 1). Create the zones first; each device then picks one."
	descInterfaces lang.Text = "The network connections of each device: its MAC and IP address. Staff log in from the devices registered here."
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
//
// Ejemplo: screens := ui.Screens(caller)
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:         ID,
			Label:      Label,
			Icon:       laptop.Ref.ID(),
			Permission: devicemanager.ResourceDevice,
			Presenter:  devicemanager.NewView(caller, view.WithDescription(descDevices)),
			Lookups: []view.Lookup{
				{Field: fieldZoneID, Source: devicemanager.NewZoneView(caller)},
			},
		},
		{
			ID:         ZonesID,
			Label:      ZonesLabel,
			Icon:       location.Ref.ID(),
			Permission: devicemanager.ResourceZone,
			Presenter:  devicemanager.NewZoneView(caller, view.WithDescription(descZones)),
		},
		{
			ID:         InterfacesID,
			Label:      InterfacesLabel,
			Icon:       network.Ref.ID(),
			Permission: devicemanager.ResourceNetworkInterface,
			Presenter:  devicemanager.NewNetworkInterfaceView(caller, view.WithDescription(descInterfaces)),
			Lookups: []view.Lookup{
				{Field: fieldDeviceID, Source: devicemanager.NewView(caller)},
			},
		},
	}
}
