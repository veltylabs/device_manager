package devicemanager

import (
	"webtyp.com/input"
	"webtyp.com/model"

	"github.com/veltylabs/device_manager/kinds"
)

// Device types, re-exported from kinds (where the literals live — ormc needs kinds in their own package).
const (
	DeviceTypeComputer = kinds.TypeComputer
	DeviceTypePrinter  = kinds.TypePrinter
	DeviceTypeServer   = kinds.TypeServer
	DeviceTypeOther    = kinds.TypeOther
)

// nameChars is the charset of human names here (devices, zones): input.Text()'s
// own set plus '-' and '_', so names copied from router comments
// ("INTERNET PC-GASTRO-NEW") validate. An explicit field whitelist replaces the
// Text kind's default floor (model.Field.Validate).
var nameChars = model.Permitted{Letters: true, Tilde: true, Numbers: true, Spaces: true,
	Extra: []rune{'.', ',', '(', ')', '-', '_'}}

func namePermitted(max int) model.Permitted {
	p := nameChars
	p.Minimum = 1
	p.Maximum = max
	return p
}

var DeviceModel = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: namePermitted(255)},
		{Name: "type", Type: kinds.DeviceType(), NotNull: true},
		{Name: "location", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Maximum: 255}},
		// Soft reference (no Ref): an empty zone_id must not become an FK violation.
		// The UI fills its options (ui/browser.go).
		{Name: "zone_id", Type: input.Select(), OmitEmpty: true},
		{Name: "access", Type: kinds.DeviceAccess(), NotNull: true},
		{Name: "is_active", Type: input.Checkbox(), NotNull: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

// Zone: an area of the site served by one access point, with an IPv4 range.
var ZoneModel = model.Definition{
	Name: "zone",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: namePermitted(64)},
		{Name: "range_start", Type: input.IP(), NotNull: true},
		{Name: "range_end", Type: input.IP(), NotNull: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

// NetworkInterface: one network card of a device. mac is optional (an
// interface without MAC still identifies its device for login by IP, but is
// not sent to the router); ip is assigned from the device's zone when empty.
var NetworkInterfaceModel = model.Definition{
	Name: "network_interface",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "device_id", Type: input.Select(), NotNull: true, Ref: &DeviceModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "label", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 32}},
		{Name: "mac", Type: input.MAC(), OmitEmpty: true},
		{Name: "ip", Type: input.IP(), OmitEmpty: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

// Transport-only (Field.DB is nil on every field) — args of the ops in ops.go.

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

var GetDeviceArgsModel = model.Definition{
	Name: "get_device_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var DeactivateDeviceArgsModel = model.Definition{
	Name: "deactivate_device_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var DeleteDeviceArgsModel = model.Definition{
	Name: "delete_device_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var ListZonesArgsModel = model.Definition{
	Name: "list_zones_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
	},
}

var DeleteZoneArgsModel = model.Definition{
	Name: "delete_zone_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var ListNetworkInterfacesArgsModel = model.Definition{
	Name: "list_network_interfaces_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "device_id", Type: model.Text()},
	},
}

var DeleteNetworkInterfaceArgsModel = model.Definition{
	Name: "delete_network_interface_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

const (
	ResourceDevice           = "device"
	ResourceZone             = "zone"
	ResourceNetworkInterface = "network_interface"
	ImportedInterfaceLabel   = "imported"
)

type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrNotFound          domainError = "device not found"
	ErrIPAlreadyExists   domainError = "device ip already exists"
	ErrMACAlreadyExists  domainError = "network interface mac already exists"
	ErrRandomizedMAC     domainError = "randomized MAC address: on the device, turn off random (private) hardware addresses for this Wi-Fi network and register its real MAC"
	ErrNoZone            domainError = "ip is empty and the device has no zone to assign one from"
	ErrIPOutsideZone     domainError = "ip is outside the device's zone range"
	ErrZoneFull          domainError = "zone has no free ip"
	ErrZoneOverlap       domainError = "zone range overlaps another zone"
	ErrZoneInUse         domainError = "zone is used by a device"
	ErrZoneNotFound      domainError = "zone not found"
	ErrInterfaceNotFound domainError = "network interface not found"
	ErrInvalidZoneRange  domainError = "zone range must be two IPv4 addresses with start <= end"
)

type ValidationError struct{ Err error }

func (v ValidationError) Error() string { return v.Err.Error() }

// <module>.<entity>.<past-tense-verb> — tenant_id goes in the event payload, never the topic name.
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
