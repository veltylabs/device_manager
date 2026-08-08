package devicemanager

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
)

const (
	OpListDevices      = "list_devices"
	OpGetDevice        = "get_device"
	OpCreateDevice     = "create_device"
	OpUpdateDevice     = "update_device"
	OpUpsertDevice     = "upsert_device"
	OpDeactivateDevice = "deactivate_device"
	OpDeleteDevice     = "delete_device"
)

func (m *Module) ModelName() string { return "device_manager" }

func (m *Module) MountOps(reg router.OpRegistry) {
	reg.Op(OpListDevices, m.opListDevices).Requires("device", model.Read).Accepts(&ListDevicesArgs{})
	reg.Op(OpGetDevice, m.opGetDevice).Requires("device", model.Read).Accepts(&GetDeviceArgs{})
	reg.Op(OpCreateDevice, m.opCreateDevice).Requires("device", model.Create).Accepts(&Device{})
	reg.Op(OpUpdateDevice, m.opUpdateDevice).Requires("device", model.Update).Accepts(&Device{})
	reg.Op(OpUpsertDevice, m.opUpsertDevice).Requires("device", model.Create|model.Update).Accepts(&Device{})
	reg.Op(OpDeactivateDevice, m.opDeactivateDevice).Requires("device", model.Update).Accepts(&DeactivateDeviceArgs{})
	reg.Op(OpDeleteDevice, m.opDeleteDevice).Requires("device", model.Delete).Accepts(&DeleteDeviceArgs{})
}

var _ router.OpModule = (*Module)(nil)

func (m *Module) opListDevices(ctx router.Context) {
	var args ListDevicesArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	filter := DeviceFilter{Type: args.Type, ActiveOnly: args.ActiveOnly, Limit: args.Limit, Offset: args.Offset}
	devices, err := m.ListDevices(args.TenantId, filter)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	list := make(DeviceList, len(devices))
	for i := range devices {
		list[i] = &devices[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opGetDevice(ctx router.Context) {
	var args GetDeviceArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	d, err := m.GetDevice(args.TenantId, args.Id)
	if err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&d); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opCreateDevice(ctx router.Context) {
	var d Device
	if err := ctx.Decode(&d); err != nil {
		ctx.WriteStatus(400)
		return
	}
	created, err := m.CreateDevice(d)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrIPAlreadyExists {
			ctx.WriteStatus(409)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&created); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpdateDevice(ctx router.Context) {
	var d Device
	if err := ctx.Decode(&d); err != nil {
		ctx.WriteStatus(400)
		return
	}
	updated, err := m.UpdateDevice(d)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&updated); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpsertDevice(ctx router.Context) {
	var d Device
	if err := ctx.Decode(&d); err != nil {
		ctx.WriteStatus(400)
		return
	}
	var out Device
	var err error
	if d.Id == "" {
		out, err = m.CreateDevice(d)
	} else {
		out, err = m.UpdateDevice(d)
	}
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrIPAlreadyExists {
			ctx.WriteStatus(409)
		} else if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&out); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opDeactivateDevice(ctx router.Context) {
	var args DeactivateDeviceArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := m.DeactivateDevice(args.TenantId, args.Id); err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	ctx.WriteStatus(200)
}

func (m *Module) opDeleteDevice(ctx router.Context) {
	var args DeleteDeviceArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := m.DeleteDevice(args.TenantId, args.Id); err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	ctx.WriteStatus(200)
}
