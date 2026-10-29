package devicemanager

import (
	"webtyp.com/input"
	"webtyp.com/model"

	"github.com/veltylabs/device_manager/kinds"
)

// Tipos de equipos reexportados desde kinds (donde residen los literales; ormc requiere kinds en su propio paquete).
//
// Ejemplo: tipo := devicemanager.DeviceTypeComputer
const (
	DeviceTypeComputer = kinds.TypeComputer
	DeviceTypePrinter  = kinds.TypePrinter
	DeviceTypeServer   = kinds.TypeServer
	DeviceTypeOther    = kinds.TypeOther
)

// nameChars es el conjunto de caracteres permitidos para nombres humanos (equipos, zonas):
// el conjunto de input.Text() más '-' y '_', de modo que nombres copiados de comentarios de router
// ("INTERNET PC-GASTRO-NEW") sean válidos. Una lista blanca explícita reemplaza el umbral
// por defecto del kind Text (model.Field.Validate).
var nameChars = model.Permitted{Letters: true, Tilde: true, Numbers: true, Spaces: true,
	Extra: []rune{'.', ',', '(', ')', '-', '_'}}

func namePermitted(max int) model.Permitted {
	p := nameChars
	p.Minimum = 1
	p.Maximum = max
	return p
}

// DeviceModel define el esquema y validaciones de un equipo en el inventario.
//
// Ejemplo: def := devicemanager.DeviceModel
var DeviceModel = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), Label: "Name", NotNull: true, Permitted: namePermitted(255), Item: model.ItemLabel},
		{Name: "type", Type: kinds.DeviceType(), Label: "Type", NotNull: true},
		{Name: "location", Type: input.Text(), Label: "Location", OmitEmpty: true, Permitted: model.Permitted{Maximum: 255}, Item: model.ItemDescription},
		// Referencia blanda (sin Ref): un zone_id vacío no debe ser una violación de FK.
		// La UI llena sus opciones (ui/browser.go).
		{Name: "zone_id", Type: input.Select(), Label: "Zone", OmitEmpty: true, Options: &ZoneModel},
		{Name: "access", Type: kinds.DeviceAccess(), Label: "Access", NotNull: true, Item: model.ItemDescription},
		{Name: "is_active", Type: input.Checkbox(), Label: "Active", NotNull: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

// ZoneModel define el esquema y validaciones de una zona del recinto (área servida por un AP con rango IPv4).
//
// Ejemplo: def := devicemanager.ZoneModel
var ZoneModel = model.Definition{
	Name: "zone",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), Label: "Name", NotNull: true, Permitted: namePermitted(64), Item: model.ItemLabel},
		{Name: "range_start", Type: input.IP(), Label: "Range start", NotNull: true, Item: model.ItemDescription},
		{Name: "range_end", Type: input.IP(), Label: "Range end", NotNull: true, Item: model.ItemDescription},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

// NetworkInterfaceModel define el esquema de una tarjeta de red de un equipo. mac es opcional;
// ip se asigna automáticamente desde la zona del equipo si está vacía.
//
// Ejemplo: def := devicemanager.NetworkInterfaceModel
var NetworkInterfaceModel = model.Definition{
	Name: "network_interface",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "device_id", Type: input.Select(), Label: "Device", NotNull: true, Ref: &DeviceModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "label", Type: input.Text(), Label: "Label", NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 32}, Item: model.ItemLabel},
		{Name: "mac", Type: input.MAC(), Label: "MAC address", OmitEmpty: true, Item: model.ItemDescription},
		{Name: "ip", Type: input.IP(), Label: "IP address", OmitEmpty: true, Item: model.ItemDescription},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

// Modelos de solo transporte (Field.DB es nil en cada campo) — argumentos de las operaciones en ops.go.

// ListDevicesArgsModel define los argumentos para listar equipos.
//
// Ejemplo: def := devicemanager.ListDevicesArgsModel
var ListDevicesArgsModel = model.Definition{
	Name: "list_devices_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "type", Type: model.Text()},
		{Name: "active_only", Type: model.Bool()},
		{Name: "limit", Type: model.Int()},
		{Name: "offset", Type: model.Int()},
	},
}

// GetDeviceArgsModel define los argumentos para obtener un equipo por ID.
//
// Ejemplo: def := devicemanager.GetDeviceArgsModel
var GetDeviceArgsModel = model.Definition{
	Name: "get_device_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

// DeactivateDeviceArgsModel define los argumentos para desactivar un equipo.
//
// Ejemplo: def := devicemanager.DeactivateDeviceArgsModel
var DeactivateDeviceArgsModel = model.Definition{
	Name: "deactivate_device_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

// DeleteDeviceArgsModel define los argumentos para eliminar un equipo.
//
// Ejemplo: def := devicemanager.DeleteDeviceArgsModel
var DeleteDeviceArgsModel = model.Definition{
	Name: "delete_device_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

// ListZonesArgsModel define los argumentos para listar zonas.
//
// Ejemplo: def := devicemanager.ListZonesArgsModel
var ListZonesArgsModel = model.Definition{
	Name: "list_zones_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
	},
}

// DeleteZoneArgsModel define los argumentos para eliminar una zona.
//
// Ejemplo: def := devicemanager.DeleteZoneArgsModel
var DeleteZoneArgsModel = model.Definition{
	Name: "delete_zone_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

// ListNetworkInterfacesArgsModel define los argumentos para listar interfaces de red.
//
// Ejemplo: def := devicemanager.ListNetworkInterfacesArgsModel
var ListNetworkInterfacesArgsModel = model.Definition{
	Name: "list_network_interfaces_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "device_id", Type: model.Text()},
	},
}

// DeleteNetworkInterfaceArgsModel define los argumentos para eliminar una interfaz de red.
//
// Ejemplo: def := devicemanager.DeleteNetworkInterfaceArgsModel
var DeleteNetworkInterfaceArgsModel = model.Definition{
	Name: "delete_network_interface_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

// Constantes de nombres de recursos y etiquetas.
//
// Constantes de etiquetas.
const (
	ImportedInterfaceLabel   = "imported"
)

// Errores de dominio del módulo.
//
// Ejemplo: if model.KindOf(err) == model.ErrorNotFound { ... }
const (
	ErrNotFound          model.NotFound = "device not found"
	ErrIPAlreadyExists   model.Conflict = "device ip already exists"
	ErrMACAlreadyExists  model.Conflict = "network interface mac already exists"
	ErrRandomizedMAC     model.Invalid  = "randomized MAC address: on the device, turn off random (private) hardware addresses for this Wi-Fi network and register its real MAC"
	ErrNoZone            model.Invalid  = "ip is empty and the device has no zone to assign one from"
	ErrIPOutsideZone     model.Invalid  = "ip is outside the device's zone range"
	ErrZoneFull          model.Conflict = "zone has no free ip"
	ErrZoneOverlap       model.Conflict = "zone range overlaps another zone"
	ErrZoneInUse         model.Conflict = "zone is used by a device"
	ErrZoneNotFound      model.NotFound = "zone not found"
	ErrInterfaceNotFound model.NotFound = "network interface not found"
	ErrInvalidZoneRange  model.Invalid  = "zone range must be two IPv4 addresses with start <= end"
	ErrCannotRead        model.Invalid  = "device_manager: cannot read the request"
)

//

// Tópicos de eventos publicados por el módulo. El tenant_id viaja en el payload del evento, nunca en el tópico.
//
// Ejemplo: topico := devicemanager.TopicDeviceCreated
const (
	TopicDeviceCreated     = "device_manager.device.created"
	TopicDeviceUpdated     = "device_manager.device.updated"
	TopicDeviceDeactivated = "device_manager.device.deactivated"
	TopicDeviceDeleted     = "device_manager.device.deleted"

	TopicZoneSaved               = "device_manager.zone.saved"
	TopicZoneDeleted             = "device_manager.zone.deleted"
	TopicNetworkInterfaceSaved   = "device_manager.network_interface.saved"
	TopicNetworkInterfaceDeleted = "device_manager.network_interface.deleted"
)
