package devicemanager

import (
	"github.com/tinywasm/ddl"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/time"
)

// Deps are the module's infrastructure ports — never a concrete implementation.
type Deps struct {
	IDs       model.IDGenerator // required — the module never builds its own
	Publisher events.Publisher  // optional — nil disables publishing silently
}

type Module struct {
	db  *orm.DB
	ids model.IDGenerator
	pub events.Publisher
}

// New connects the module to an already-connected *orm.DB and migrates its own schema when the
// backend supports DDL (storage/mem, used by this module's own tests, does not — the type
// assertion below is how the module stays agnostic to that, same idiom as every sibling module).
func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("device_manager: Deps.IDs is required")
	}
	if ddlCompiler, ok := db.RawConn().(ddl.Compiler); ok {
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&Device{}); err != nil {
			return nil, err
		}
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}

// DeviceFilter narrows ListDevices — a plain internal type, never ormc-generated (only its
// transport twin ListDevicesArgs, in model_orm.go, is).
type DeviceFilter struct {
	Type       string
	ActiveOnly bool
	Limit      int64
	Offset     int64
}

func (m *Module) GetDevice(tenantId, id string) (Device, error) {
	var d Device
	qb := m.db.Query(&d).Where(Device_.Id).Eq(id).Where(Device_.TenantId).Eq(tenantId)
	_, err := ReadOneDevice(qb, &d)
	if err != nil {
		// Never hide a real DB failure as "not found" — only orm.ErrNotFound maps to the domain
		// sentinel; anything else surfaces as the internal error it is.
		if err == orm.ErrNotFound {
			return Device{}, ErrNotFound
		}
		return Device{}, err
	}
	return d, nil
}

func (m *Module) FindByIP(tenantId, ip string) (Device, error) {
	var d Device
	qb := m.db.Query(&d).Where(Device_.Ip).Eq(ip).Where(Device_.TenantId).Eq(tenantId)
	_, err := ReadOneDevice(qb, &d)
	if err != nil {
		if err == orm.ErrNotFound {
			return Device{}, ErrNotFound
		}
		return Device{}, err
	}
	return d, nil
}

func (m *Module) ListDevices(tenantId string, filter DeviceFilter) ([]Device, error) {
	var d Device
	qb := m.db.Query(&d).Where(Device_.TenantId).Eq(tenantId)
	if filter.Type != "" {
		qb = qb.Where(Device_.Type).Eq(filter.Type)
	}
	if filter.ActiveOnly {
		qb = qb.Where(Device_.IsActive).Eq(true)
	}
	if filter.Limit > 0 {
		qb = qb.Limit(int(filter.Limit))
	}
	if filter.Offset > 0 {
		qb = qb.Offset(int(filter.Offset))
	}
	results, err := ReadAllDevice(qb)
	if err != nil {
		return nil, err
	}
	devices := make([]Device, len(results))
	for i, r := range results {
		devices[i] = *r
	}
	return devices, nil
}

// isValidDeviceType is the single place the 4 DeviceType* constants are checked against — never
// duplicate this switch/comparison at another call site.
func isValidDeviceType(t string) bool {
	return t == DeviceTypeComputer || t == DeviceTypePrinter || t == DeviceTypeServer || t == DeviceTypeOther
}

func (m *Module) CreateDevice(d Device) (Device, error) {
	d.Id = m.ids.NewID()
	d.UpdatedAt = time.Now()

	if err := d.Validate(model.ActionCreate); err != nil {
		return Device{}, ValidationError{Err: err}
	}
	if !isValidDeviceType(d.Type) {
		return Device{}, ValidationError{Err: fmt.Err("invalid device type: %s", d.Type)}
	}

	// IP uniqueness is enforced per tenant, not globally — two tenants may reuse the same
	// private IP range independently.
	existing, err := m.FindByIP(d.TenantId, d.Ip)
	if err == nil {
		if existing.Id != "" {
			return Device{}, ErrIPAlreadyExists
		}
	} else if err != ErrNotFound {
		return Device{}, err
	}

	if err := m.db.Create(&d); err != nil {
		return Device{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceCreated, Payload: &d})
	}
	return d, nil
}

func (m *Module) UpdateDevice(d Device) (Device, error) {
	if err := d.Validate(model.ActionUpdate); err != nil {
		return Device{}, ValidationError{Err: err}
	}
	if !isValidDeviceType(d.Type) {
		return Device{}, ValidationError{Err: fmt.Err("invalid device type: %s", d.Type)}
	}

	// Verify the device exists and belongs to this tenant before writing.
	if _, err := m.GetDevice(d.TenantId, d.Id); err != nil {
		return Device{}, err
	}

	d.UpdatedAt = time.Now()
	if err := m.db.Update(&d, orm.Eq(Device_.Id, d.Id), orm.Eq(Device_.TenantId, d.TenantId)); err != nil {
		return Device{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceUpdated, Payload: &d})
	}
	return d, nil
}

func (m *Module) DeactivateDevice(tenantId, id string) error {
	d, err := m.GetDevice(tenantId, id)
	if err != nil {
		return err
	}
	d.IsActive = false
	d.UpdatedAt = time.Now()
	if err := m.db.Update(&d, orm.Eq(Device_.Id, d.Id), orm.Eq(Device_.TenantId, d.TenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceDeactivated, Payload: &d})
	}
	return nil
}

func (m *Module) DeleteDevice(tenantId, id string) error {
	d, err := m.GetDevice(tenantId, id)
	if err != nil {
		return err
	}
	if err := m.db.Delete(&d, orm.Eq(Device_.Id, d.Id), orm.Eq(Device_.TenantId, d.TenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceDeleted, Payload: &d})
	}
	return nil
}
