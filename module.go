package devicemanager

import (
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/lang"
	"webtyp.com/model"
	"webtyp.com/orm"
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
	store    *orm.Store
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
	return &Module{store: orm.TenantStore(db, deps.TenantID, deps.IDs), db: db, ids: deps.IDs, pub: deps.Publisher, tenantID: deps.TenantID}, nil
}

// DeviceFilter acota los resultados de ListDevices — un tipo interno directo, nunca generado por ormc
// (su gemelo de transporte es ListDevicesArgs en model_orm.go).
//
// Ejemplo: f := devicemanager.DeviceFilter{ActiveOnly: true}
type DeviceFilter struct {
	Search     string
	After      string
	Type       string
	ActiveOnly bool
	Limit      int64
	Offset     int64
}

// GetDevice obtiene un equipo por su ID para un tenant específico.
//
// Ejemplo: d, err := dm.GetDevice(tenantID, deviceID)
func (m *Module) GetDevice(tenantId, id string) (Device, error) {
	if tenantId != m.tenantID {
		return Device{}, nil
	}
	var d Device
	qb := m.db.Query(&d).Where(Device_.Id).Eq(id)
	if _, err := ReadOneDevice(qb, &d); err != nil {
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
	if tenantId != m.tenantID {
		return Device{}, nil
	}
	var ni NetworkInterface
	qb := m.db.Query(&ni).Where(NetworkInterface_.Ip).Eq(input.CanonicalIP(ip))
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
	if tenantId != m.tenantID {
		return nil, nil
	}
	var d Device
	qb := m.db.Query(&d).Where(Device_.TenantId).Eq(tenantId)
	if filter.Search != "" {
		qb = qb.Where(Device_.Name).Like("%" + filter.Search + "%")
	}
	if filter.After != "" {
		qb = qb.Where(Device_.Id).Gt(filter.After)
	}
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
	qb = qb.OrderBy(Device_.Id).Asc()

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





