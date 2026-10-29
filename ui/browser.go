package ui

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/lang"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
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
			Permission: model.ResourceOf(&devicemanager.DeviceModel),
			Presenter:  devicemanager.NewView(caller, view.WithDescription(descDevices)),
		},
		{
			ID:         ZonesID,
			Label:      ZonesLabel,
			Icon:       location.Ref.ID(),
			Permission: model.ResourceOf(&devicemanager.ZoneModel),
			Presenter:  devicemanager.NewZoneView(caller, view.WithDescription(descZones)),
		},
		{
			ID:         InterfacesID,
			Label:      InterfacesLabel,
			Icon:       network.Ref.ID(),
			Permission: model.ResourceOf(&devicemanager.NetworkInterfaceModel),
			Presenter:  devicemanager.NewNetworkInterfaceView(caller, view.WithDescription(descInterfaces)),
		},
	}
}
