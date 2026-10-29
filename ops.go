package devicemanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/rpc"
	"webtyp.com/view"
)

// Nombres de operaciones expuestas por el módulo.
//
// Ejemplo: op := devicemanager.OpListDevices
const (
	OpListDevices      = "list_devices"
	OpGetDevice        = "get_device"
	OpCreateDevice     = "create_device"
	OpUpdateDevice     = "update_device"
	OpUpsertDevice     = "upsert_device"
	OpDeactivateDevice = "deactivate_device"
	OpDeleteDevice     = "delete_device"

	OpListZones              = "list_zones"
	OpSaveZone               = "save_zone"
	OpDeleteZone             = "delete_zone"
	OpListNetworkInterfaces  = "list_network_interfaces"
	OpSaveNetworkInterface   = "save_network_interface"
	OpDeleteNetworkInterface = "delete_network_interface"
)

// ModelName es la identidad de este módulo: mcp.HarvestOps califica cada operación
// como "device_manager.<nombre>" en el transporte, y NewView en view.go usa esta
// misma constante como view.Ops.Module para que el cliente genere nombres idénticos.
//
// Ejemplo: name := devicemanager.ModelName
const ModelName = "device_manager"

// ModelName devuelve el identificador de módulo en el router.
//
// Ejemplo: name := dm.ModelName()
func (m *Module) ModelName() string { return ModelName }

// MountOperations registra las operaciones del módulo en el registro de operaciones del router.
//
// Ejemplo: dm.MountOperations(reg)
func (m *Module) saveDevice(d *Device) error {
	d.TenantId = m.tenantOr(d.TenantId)
	var out Device
	var err error
	if d.Id == "" {
		out, err = m.CreateDevice(*d)
	} else {
		out, err = m.UpdateDevice(*d)
	}
	*d = out
	return err
}

func (m *Module) MountOperations(reg router.OperationRegistry) {
	reg.Operation(OpListDevices, m.opListDevices).Requires(model.ResourceOf(&DeviceModel), model.Read).Accepts(&ListDevicesArgs{})
	reg.Operation(OpGetDevice, m.opGetDevice).Requires(model.ResourceOf(&DeviceModel), model.Read).Accepts(&GetDeviceArgs{})
	reg.Operation(OpCreateDevice, m.opCreateDevice).Requires(model.ResourceOf(&DeviceModel), model.Create).Accepts(&Device{})
	reg.Operation(OpUpdateDevice, m.opUpdateDevice).Requires(model.ResourceOf(&DeviceModel), model.Update).Accepts(&Device{})
	view.MountSave(reg, OpUpsertDevice, m.saveDevice).Requires(model.ResourceOf(&DeviceModel), model.Create|model.Update)
	reg.Operation(OpDeactivateDevice, m.opDeactivateDevice).Requires(model.ResourceOf(&DeviceModel), model.Update).Accepts(&DeactivateDeviceArgs{})
	view.MountDelete(reg, OpDeleteDevice, func(ids []string) error {
		for _, id := range ids {
			if err := m.DeleteDevice(m.tenantID, id); err != nil {
				return err
			}
		}
		return nil
	}).Requires(model.ResourceOf(&DeviceModel), model.Delete)
	reg.Operation(OpListZones, m.opListZones).Requires(model.ResourceOf(&ZoneModel), model.Read).Accepts(&ListZonesArgs{})
	view.MountSave(reg, OpSaveZone, func(z *Zone) error {
		z.TenantId = m.tenantOr(z.TenantId)
		out, err := m.SaveZone(*z)
		*z = out
		return err
	}).Requires(model.ResourceOf(&ZoneModel), model.Create|model.Update)
	view.MountDelete(reg, OpDeleteZone, func(ids []string) error {
		for _, id := range ids {
			if err := m.DeleteZone(m.tenantID, id); err != nil {
				return err
			}
		}
		return nil
	}).Requires(model.ResourceOf(&ZoneModel), model.Delete)
	reg.Operation(OpListNetworkInterfaces, m.opListNetworkInterfaces).Requires(model.ResourceOf(&NetworkInterfaceModel), model.Read).Accepts(&ListNetworkInterfacesArgs{})
	view.MountSave(reg, OpSaveNetworkInterface, func(ni *NetworkInterface) error {
		ni.TenantId = m.tenantOr(ni.TenantId)
		out, err := m.SaveNetworkInterface(*ni)
		*ni = out
		return err
	}).Requires(model.ResourceOf(&NetworkInterfaceModel), model.Create|model.Update)
	view.MountDelete(reg, OpDeleteNetworkInterface, func(ids []string) error {
		for _, id := range ids {
			if err := m.DeleteNetworkInterface(m.tenantID, id); err != nil {
				return err
			}
		}
		return nil
	}).Requires(model.ResourceOf(&NetworkInterfaceModel), model.Delete)
}

var _ router.OperationModule = (*Module)(nil)

func (m *Module) opListDevices(ctx router.Context) {
	var args ListDevicesArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	tenantID := args.TenantId
	if tenantID == "" {
		tenantID = m.tenantID
	}
	filter := DeviceFilter{Type: args.Type, ActiveOnly: args.ActiveOnly, Limit: args.Limit, Offset: args.Offset}
	devices, err := m.ListDevices(tenantID, filter)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(DeviceList, len(devices))
	for i := range devices {
		list[i] = &devices[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opGetDevice(ctx router.Context) {
	var args GetDeviceArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	d, err := m.GetDevice(args.TenantId, args.Id)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(&d); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opCreateDevice(ctx router.Context) {
	var d Device
	if err := ctx.Decode(&d); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	created, err := m.CreateDevice(d)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(&created); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opUpdateDevice(ctx router.Context) {
	var d Device
	if err := ctx.Decode(&d); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	updated, err := m.UpdateDevice(d)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	if err := ctx.Encode(&updated); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opDeactivateDevice(ctx router.Context) {
	var args DeactivateDeviceArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	if err := m.DeactivateDevice(args.TenantId, args.Id); err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	ctx.WriteStatus(200)
}

// tenantOr returns id, or the installation's tenant when the caller sent none
// (every crudview-backed form does) — same fallback as opListDevices.
func (m *Module) tenantOr(id string) string {
	if id == "" {
		return m.tenantID
	}
	return id
}

func (m *Module) opListZones(ctx router.Context) {
	var args ListZonesArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	zones, err := m.ListZones(m.tenantOr(args.TenantId))
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(ZoneList, len(zones))
	for i := range zones {
		list[i] = &zones[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

func (m *Module) opListNetworkInterfaces(ctx router.Context) {
	var args ListNetworkInterfacesArgs
	if err := ctx.Decode(&args); err != nil {
		rpc.WriteError(ctx, ErrCannotRead)
		return
	}
	ifaces, err := m.ListNetworkInterfaces(m.tenantOr(args.TenantId), args.DeviceId)
	if err != nil {
		rpc.WriteError(ctx, err)
		return
	}
	list := make(NetworkInterfaceList, len(ifaces))
	for i := range ifaces {
		list[i] = &ifaces[i]
	}
	if err := ctx.Encode(&list); err != nil {
		rpc.WriteError(ctx, err)
	}
}

