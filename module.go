package devicemanager

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/lang"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/orm"
	"webtyp.com/time"
)

// Deps define los puertos de infraestructura del módulo — nunca una implementación concreta.
//
// Ejemplo: deps := devicemanager.Deps{IDs: idGen, TenantID: "tenant-1"}
type Deps struct {
	IDs       model.IDGenerator // requerido — el módulo nunca genera sus propios IDs
	Publisher events.Publisher  // opcional — nil desactiva la publicación silenciosamente
	TenantID  string            // requerido — tenant id de esta instalación, valor por defecto
	// cuando el invocador no envía tenant_id (típico en listas de crudview)
}

// Module es el servicio principal del inventario de equipos.
//
// Ejemplo: dm, err := devicemanager.New(db, deps)
type Module struct {
	db       *orm.DB
	ids      model.IDGenerator
	pub      events.Publisher
	tenantID string
}

// New conecta el módulo a una instancia ya conectada de *orm.DB; el esquema se asume
// preexistente — ver el subpaquete migrate.
//
// Ejemplo: dm, err := devicemanager.New(db, deps)
// Nombres de las dependencias que New exige: son dato (código), no texto a traducir;
// por eso son variables y no constantes (langc traduce las constantes que recibe fmt.Err).
var (
	errDepsIDs      = "device_manager: Deps.IDs"
	errDepsTenantID = "device_manager: Deps.TenantID"
)

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err(errDepsIDs, lang.Text("is required"))
	}
	if deps.TenantID == "" {
		return nil, fmt.Err(errDepsTenantID, lang.Text("is required"))
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher, tenantID: deps.TenantID}, nil
}

// DeviceFilter acota los resultados de ListDevices — un tipo interno directo, nunca generado por ormc
// (su gemelo de transporte es ListDevicesArgs en model_orm.go).
//
// Ejemplo: f := devicemanager.DeviceFilter{ActiveOnly: true}
type DeviceFilter struct {
	Type       string
	ActiveOnly bool
	Limit      int64
	Offset     int64
}

// GetDevice obtiene un equipo por su ID para un tenant específico.
//
// Ejemplo: d, err := dm.GetDevice(tenantID, deviceID)
func (m *Module) GetDevice(tenantId, id string) (Device, error) {
	var d Device
	qb := m.db.Query(&d).Where(Device_.Id).Eq(id).Where(Device_.TenantId).Eq(tenantId)
	_, err := ReadOneDevice(qb, &d)
	if err != nil {
		// Nunca ocultar un fallo real de BD como "not found" — solo orm.ErrNotFound se mapea al
		// centinela de dominio; cualquier otro error sube como el error interno que es.
		if orm.IsNotFound(err) {
			return Device{}, ErrNotFound
		}
		return Device{}, err
	}
	return d, nil
}

// FindByIP busca la IP en su formato canónico (input.CanonicalIP) entre las interfaces
// de red del tenant y retorna el equipo dueño — el formato en que se almacenan las interfaces
// y el formato en que reporta ClientIP de webtyp.com/auth.
//
// Ejemplo: d, err := dm.FindByIP(tenantID, "172.0.0.65")
func (m *Module) FindByIP(tenantId, ip string) (Device, error) {
	var ni NetworkInterface
	qb := m.db.Query(&ni).Where(NetworkInterface_.Ip).Eq(input.CanonicalIP(ip)).Where(NetworkInterface_.TenantId).Eq(tenantId)
	if _, err := ReadOneNetworkInterface(qb, &ni); err != nil {
		if orm.IsNotFound(err) {
			return Device{}, ErrNotFound
		}
		return Device{}, err
	}
	return m.GetDevice(tenantId, ni.DeviceId)
}

// ListDevices retorna los equipos de un tenant según el filtro especificado.
//
// Ejemplo: devices, err := dm.ListDevices(tenantID, filter)
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

// isValidDeviceType es el único lugar donde se verifican las 4 constantes DeviceType*.
func isValidDeviceType(t string) bool {
	return t == DeviceTypeComputer || t == DeviceTypePrinter || t == DeviceTypeServer || t == DeviceTypeOther
}

// validateDevice ejecuta todas las validaciones de escritura para un equipo.
func (m *Module) validateDevice(d Device, action byte) error {
	if err := d.Validate(action); err != nil {
		return err
	}
	if !isValidDeviceType(d.Type) {
		return &model.FieldError{Field: "type", Err: fmt.Err("invalid device type: %s", d.Type)}
	}
	if _, err := network.ParseAccess(d.Access); err != nil {
		return err
	}
	if d.ZoneId != "" {
		if _, err := m.GetZone(d.TenantId, d.ZoneId); err != nil {
			return &model.FieldError{Field: "zone_id", Err: err}
		}
	}
	return nil
}

// CreateDevice crea un nuevo equipo en el inventario.
//
// Ejemplo: nuevo, err := dm.CreateDevice(d)
func (m *Module) CreateDevice(d Device) (Device, error) {
	d.Id = m.ids.NewID()
	if d.Access == "" {
		d.Access = network.AccessLocalName
	}
	d.UpdatedAt = time.NowNano()
	if err := m.validateDevice(d, model.ActionCreate); err != nil {
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

// UpdateDevice actualiza un equipo existente en el inventario.
//
// Ejemplo: actualizado, err := dm.UpdateDevice(d)
func (m *Module) UpdateDevice(d Device) (Device, error) {
	if err := m.validateDevice(d, model.ActionUpdate); err != nil {
		return Device{}, err
	}
	// Verificar que el equipo exista y pertenezca a este tenant antes de actualizar.
	if _, err := m.GetDevice(d.TenantId, d.Id); err != nil {
		return Device{}, err
	}
	d.UpdatedAt = time.NowNano()
	if err := m.db.Update(&d, orm.Eq(Device_.Id, d.Id), orm.Eq(Device_.TenantId, d.TenantId)); err != nil {
		return Device{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceUpdated, Payload: &d})
	}
	return d, nil
}

// DeactivateDevice desactiva lógicamente un equipo (is_active = false).
//
// Ejemplo: err := dm.DeactivateDevice(tenantID, deviceID)
func (m *Module) DeactivateDevice(tenantId, id string) error {
	d, err := m.GetDevice(tenantId, id)
	if err != nil {
		return err
	}
	d.IsActive = false
	d.UpdatedAt = time.NowNano()
	if err := m.db.Update(&d, orm.Eq(Device_.Id, d.Id), orm.Eq(Device_.TenantId, d.TenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceDeactivated, Payload: &d})
	}
	return nil
}

// DeleteDevice elimina definitivamente un equipo y todas sus interfaces asociadas.
//
// Ejemplo: err := dm.DeleteDevice(tenantID, deviceID)
func (m *Module) DeleteDevice(tenantId, id string) error {
	d, err := m.GetDevice(tenantId, id)
	if err != nil {
		return err
	}
	// Primero sus interfaces: network_interface.device_id referencia a device.id.
	ifaces, err := m.ListNetworkInterfaces(tenantId, id)
	if err != nil {
		return err
	}
	for _, ni := range ifaces {
		if err := m.DeleteNetworkInterface(tenantId, ni.Id); err != nil {
			return err
		}
	}
	if err := m.db.Delete(&d, orm.Eq(Device_.Id, d.Id), orm.Eq(Device_.TenantId, d.TenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicDeviceDeleted, Payload: &d})
	}
	return nil
}
