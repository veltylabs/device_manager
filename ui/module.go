package ui

// ID es la identidad de este módulo: prefijo de recurso RBAC en el servidor y
// ruta de navegación en el cliente.
//
// Ejemplo: moduleID := ui.ID
const ID = "device_manager"

// Label es el texto para mostrar en el menú de navegación.
//
// Ejemplo: label := ui.Label
const Label = "Equipos"

// Identificadores y etiquetas de las pantallas de zonas e interfaces de red.
//
// Ejemplo: zonesID := ui.ZonesID
const (
	ZonesID         = "device_manager_zones"
	ZonesLabel      = "Zonas"
	InterfacesID    = "device_manager_interfaces"
	InterfacesLabel = "Interfaces de red"
)
