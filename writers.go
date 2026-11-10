package devicemanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/router"
)

// writerUser es quien firma las escrituras que el propio módulo hace por sí mismo (importadores, semillas, otros módulos).
const writerUser = "importer"

// write ejecuta una operación de este módulo por su propio loopback: así las reglas de los archivos rule_*.go son el
// único camino que escribe, y los métodos de abajo son solo envoltorios.
func (m *Module) write(entity string, action model.Action, args model.Encodable, into model.Decodable) error {
	var callErr error
	crud.LoopbackActingAs(m.pub, writerUser, m).Call(
		router.Operation{Module: ModelName, Entity: entity, Action: action},
		args, into, func(err error) { callErr = err })
	return callErr
}

// editableFields responde las columnas que una actualización puede enviar: todas menos la clave y las del servidor.
func editableFields(def model.Definition) []string {
	var names []string
	for _, field := range def.Fields {
		if field.ServerOnly || field.IsPK() || field.Name == "id" {
			continue
		}
		names = append(names, field.Name)
	}
	return names
}

// otherTenant dice si tenantId pide otro inquilino que el de este módulo.
func (m *Module) otherTenant(tenantId string) bool { return tenantId != "" && tenantId != m.tenantID }

// CreateDevice crea un nuevo equipo en el inventario.
//
// Ejemplo: nuevo, err := dm.CreateDevice(d)
func (m *Module) CreateDevice(d Device) (Device, error) {
	if m.otherTenant(d.TenantId) {
		return Device{}, ErrNotFound
	}
	d.Id, d.TenantId, d.UpdatedAt = "", "", 0
	if d.Access == "" {
		d.Access = network.AccessLocalName
	}
	var saved Device
	if err := m.write("device", model.Create, &crud.CreateArgs{Records: []model.Model{&d}}, &saved); err != nil {
		return Device{}, err
	}
	return saved, nil
}

// UpdateDevice actualiza un equipo existente en el inventario.
//
// Ejemplo: actualizado, err := dm.UpdateDevice(d)
func (m *Module) UpdateDevice(d Device) (Device, error) {
	if m.otherTenant(d.TenantId) {
		return Device{}, ErrNotFound
	}
	id := d.Id
	if _, err := m.GetDevice(m.tenantID, id); err != nil {
		return Device{}, err
	}
	d.Id, d.TenantId, d.UpdatedAt = "", "", 0
	err := m.write("device", model.Update,
		&crud.UpdateArgs{IDs: []string{id}, Fields: editableFields(DeviceModel), Record: &d}, nil)
	if err != nil {
		return Device{}, err
	}
	return m.GetDevice(m.tenantID, id)
}

// DeactivateDevice desactiva lógicamente un equipo (is_active = false).
//
// Ejemplo: err := dm.DeactivateDevice(tenantID, deviceID)
func (m *Module) DeactivateDevice(tenantId, id string) error {
	if m.otherTenant(tenantId) {
		return ErrNotFound
	}
	if _, err := m.GetDevice(m.tenantID, id); err != nil {
		return err
	}
	return m.write("device", model.Update,
		&crud.UpdateArgs{IDs: []string{id}, Fields: []string{Device_.IsActive}, Record: &Device{IsActive: false}}, nil)
}

// DeleteDevice elimina definitivamente un equipo y todas sus interfaces asociadas.
//
// Ejemplo: err := dm.DeleteDevice(tenantID, deviceID)
func (m *Module) DeleteDevice(tenantId, id string) error {
	if m.otherTenant(tenantId) {
		return ErrNotFound
	}
	if _, err := m.GetDevice(m.tenantID, id); err != nil {
		return err
	}
	return m.write("device", model.Delete, &crud.DeleteArgs{IDs: []string{id}}, nil)
}

// SaveZone crea la zona si Id está vacío, o la actualiza si ya existe.
//
// Ejemplo: guardada, err := dm.SaveZone(z)
func (m *Module) SaveZone(z Zone) (Zone, error) {
	if m.otherTenant(z.TenantId) {
		return Zone{}, ErrNotFound
	}
	id := z.Id
	z.Id, z.TenantId, z.UpdatedAt = "", "", 0
	if id == "" {
		var saved Zone
		if err := m.write("zone", model.Create, &crud.CreateArgs{Records: []model.Model{&z}}, &saved); err != nil {
			return Zone{}, err
		}
		return saved, nil
	}
	err := m.write("zone", model.Update,
		&crud.UpdateArgs{IDs: []string{id}, Fields: editableFields(ZoneModel), Record: &z}, nil)
	if err != nil {
		return Zone{}, err
	}
	return m.GetZone(m.tenantID, id)
}

// DeleteZone elimina una zona que no esté en uso por ningún equipo.
//
// Ejemplo: err := dm.DeleteZone(tenantID, zoneID)
func (m *Module) DeleteZone(tenantId, id string) error {
	if m.otherTenant(tenantId) {
		return ErrNotFound
	}
	if _, err := m.GetZone(m.tenantID, id); err != nil {
		return err
	}
	return m.write("zone", model.Delete, &crud.DeleteArgs{IDs: []string{id}}, nil)
}

// SaveNetworkInterface crea la interfaz si Id está vacío, o la actualiza si ya existe.
//
// Ejemplo: guardada, err := dm.SaveNetworkInterface(ni)
func (m *Module) SaveNetworkInterface(ni NetworkInterface) (NetworkInterface, error) {
	if m.otherTenant(ni.TenantId) {
		return NetworkInterface{}, ErrNotFound
	}
	id := ni.Id
	ni.Id, ni.TenantId, ni.UpdatedAt = "", "", 0
	if id == "" {
		var saved NetworkInterface
		if err := m.write("network_interface", model.Create, &crud.CreateArgs{Records: []model.Model{&ni}}, &saved); err != nil {
			return NetworkInterface{}, err
		}
		return saved, nil
	}
	err := m.write("network_interface", model.Update,
		&crud.UpdateArgs{IDs: []string{id}, Fields: editableFields(NetworkInterfaceModel), Record: &ni}, nil)
	if err != nil {
		return NetworkInterface{}, err
	}
	return m.GetNetworkInterface(m.tenantID, id)
}

// DeleteNetworkInterface elimina una interfaz de red.
//
// Ejemplo: err := dm.DeleteNetworkInterface(tenantID, interfaceID)
func (m *Module) DeleteNetworkInterface(tenantId, id string) error {
	if m.otherTenant(tenantId) {
		return ErrNotFound
	}
	if _, err := m.GetNetworkInterface(m.tenantID, id); err != nil {
		return err
	}
	return m.write("network_interface", model.Delete, &crud.DeleteArgs{IDs: []string{id}}, nil)
}
