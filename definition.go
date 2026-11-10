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
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Update: model.AddMyBusinessRule, Delete: model.HaveNoBusinessRule, Read: model.HaveNoBusinessRule},
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, ServerOnly: true, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "name", Type: input.Text(), Label: "Name", NotNull: true, Permitted: namePermitted(255), Item: model.ItemLabel},
		{Name: "type", Type: kinds.DeviceType(), Label: "Type", NotNull: true},
		{Name: "location", Type: input.Text(), Label: "Location", OmitEmpty: true, Permitted: model.Permitted{Maximum: 255}, Item: model.ItemDescription},
		// Referencia blanda (sin Ref): un zone_id vacío no debe ser una violación de FK.
		// La UI llena sus opciones (ui/browser.go).
		{Name: "zone_id", Type: input.Select(), Label: "Zone", OmitEmpty: true, Options: &ZoneModel},
		{Name: "access", Type: kinds.DeviceAccess(), Label: "Access", NotNull: true, Item: model.ItemDescription},
		{Name: "is_active", Type: input.Checkbox(), Label: "Active", NotNull: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

// ZoneModel define el esquema y validaciones de una zona del recinto (área servida por un AP con rango IPv4).
//
// Ejemplo: def := devicemanager.ZoneModel
var ZoneModel = model.Definition{
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Update: model.AddMyBusinessRule, Delete: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule},
	Name: "zone",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, ServerOnly: true, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "name", Type: input.Text(), Label: "Name", NotNull: true, Permitted: namePermitted(64), Item: model.ItemLabel},
		{Name: "range_start", Type: input.IP(), Label: "Range start", NotNull: true, Item: model.ItemDescription},
		{Name: "range_end", Type: input.IP(), Label: "Range end", NotNull: true, Item: model.ItemDescription},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

// NetworkInterfaceModel define el esquema de una tarjeta de red de un equipo. mac es opcional;
// ip se asigna automáticamente desde la zona del equipo si está vacía.
//
// Ejemplo: def := devicemanager.NetworkInterfaceModel
var NetworkInterfaceModel = model.Definition{
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Update: model.AddMyBusinessRule, Delete: model.HaveNoBusinessRule, Read: model.HaveNoBusinessRule},
	Name: "network_interface",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, ServerOnly: true, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "device_id", Type: input.Select(), Label: "Device", NotNull: true, Ref: &DeviceModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "label", Type: input.Text(), Label: "Label", NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 32}, Item: model.ItemLabel},
		{Name: "mac", Type: input.MAC(), Label: "MAC address", OmitEmpty: true, Item: model.ItemDescription},
		{Name: "ip", Type: input.IP(), Label: "IP address", OmitEmpty: true, Item: model.ItemDescription},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

// Modelos de solo transporte (Field.DB es nil en cada campo) — argumentos de las operaciones en ops.go.

// ListDevicesArgsModel define los argumentos para listar equipos.
// GetDeviceArgsModel define los argumentos para obtener un equipo por ID.
//
// Ejemplo: def := devicemanager.GetDeviceArgsModel
var GetDeviceArgsModel = model.Definition{
	Wire: true,
	Name: "get_device_args",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), ServerOnly: true},
	},
}

// DeactivateDeviceArgsModel define los argumentos para desactivar un equipo.
// ListZonesArgsModel define los argumentos para listar zonas.
// ListNetworkInterfacesArgsModel define los argumentos para listar interfaces de red.
// ImportedInterfaceLabel es la etiqueta de la interfaz creada al importar hosts.
//
// Ejemplo: ni.Label = devicemanager.ImportedInterfaceLabel
const ImportedInterfaceLabel = "imported"

// Errores de dominio del módulo.
//
// Ejemplo: if model.KindOf(err) == model.ErrorNotFound { ... }
const (
	ErrNotFound          model.NotFound = "device no encontrado"
	ErrIPAlreadyExists   model.Conflict = "ip duplicada"
	ErrMACAlreadyExists  model.Conflict = "mac duplicada"
	ErrRandomizedMAC     model.Invalid  = "dirección MAC aleatoria: apaga las direcciones privadas"
	ErrNoZone            model.Invalid  = "ip vacía y sin zona"
	ErrIPOutsideZone     model.Invalid  = "ip fuera de rango de zona"
	ErrZoneFull          model.Conflict = "zona sin ip libre"
	ErrZoneOverlap       model.Conflict = "rango solapado"
	ErrZoneInUse         model.Conflict = "zona en uso"
	ErrZoneNotFound      model.NotFound = "zona no encontrada"
	ErrInterfaceNotFound model.NotFound = "network interface not found"
	ErrInvalidZoneRange  model.Invalid  = "rango inválido"
)



var Definition = model.ModuleDefinition{
	Name: "device_manager",
	Entities: []model.Definition{
		DeviceModel,
		ZoneModel,
		NetworkInterfaceModel,
		GetDeviceArgsModel,
	},
}
