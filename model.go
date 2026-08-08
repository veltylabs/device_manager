package devicemanager

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/model"
)

// Exported, greppable, the ONLY place these literals exist — enum-valued column rule (AGENTS.md).
const (
	DeviceTypeComputer = "computer"
	DeviceTypePrinter  = "printer"
	DeviceTypeServer   = "server"
	DeviceTypeOther    = "other"
)

// Helpers to satisfy ormc's AST static type parser when the same Kind constructor is reused
// across more than one Definition (same defensive pattern as item_catalog/model.go).
var (
	BaseBool_FieldBool = model.Bool()
	BaseInt_FieldInt   = model.Int()
)

// deviceTypeWidget: closed-options radio widget for the `type` field — the
// form/input/gender.go pattern AGENTS.md requires for a user-editable enum.
type deviceTypeWidget struct {
	input.Base
}

func (w *deviceTypeWidget) Clone(parentID, name string) input.Input {
	clone := *w
	clone.InitBase(parentID, name, "radio")
	return &clone
}

func deviceType() input.Input {
	w := &deviceTypeWidget{}
	w.Letters = true
	w.Minimum = 1
	w.InitBase("", "", "radio")
	w.SetOptions(
		fmt.KeyValue{Key: DeviceTypeComputer, Value: "Computer"},
		fmt.KeyValue{Key: DeviceTypePrinter, Value: "Printer"},
		fmt.KeyValue{Key: DeviceTypeServer, Value: "Server"},
		fmt.KeyValue{Key: DeviceTypeOther, Value: "Other"},
	)
	return w
}

var DeviceModel = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 255}},
		{Name: "ip", Type: input.IP(), NotNull: true},
		{Name: "type", Type: deviceType(), NotNull: true},
		{Name: "location", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Maximum: 255}},
		{Name: "is_active", Type: input.Checkbox(), NotNull: true},
		{Name: "updated_at", Type: BaseInt_FieldInt, OmitEmpty: true},
	},
}

// Transport-only (Field.DB is nil on every field) — args of the ops in ops.go.

var ListDevicesArgsModel = model.Definition{
	Name: "list_devices_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "type", Type: model.Text()},
		{Name: "active_only", Type: BaseBool_FieldBool},
		{Name: "limit", Type: BaseInt_FieldInt},
		{Name: "offset", Type: BaseInt_FieldInt},
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

var (
	ErrNotFound        = fmt.Err("device not found")
	ErrIPAlreadyExists = fmt.Err("device ip already exists")
)

type ValidationError struct{ Err error }

func (v ValidationError) Error() string { return v.Err.Error() }

// <module>.<entity>.<past-tense-verb> — tenant_id goes in the event payload, never the topic name.
const (
	TopicDeviceCreated     = "device_manager.device.created"
	TopicDeviceUpdated     = "device_manager.device.updated"
	TopicDeviceDeactivated = "device_manager.device.deactivated"
	TopicDeviceDeleted     = "device_manager.device.deleted"
)
