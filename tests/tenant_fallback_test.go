package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
)

// TestOpListDevices_FallsBackToModuleTenant reproduce un bug real encontrado al
// probar manualmente la pantalla "Equipos" de una app derivada: la lista mostraba
// "0 / 0" aunque existía una fila real de equipo para el tenant de la app.
//
// Causa raíz: callerLister.list() de webtyp.com/view — el código que cada
// lista respaldada por crudview ejecuta en Reload() — siempre llama a la op List SIN
// argumentos (ver la propia documentación de router.Caller: esto es correcto, por diseño, ya que una
// op List se escribe para tolerar un objeto args ausente/vacío). opListDevices
// por lo tanto recibe ListDevicesArgs{TenantId: ""} en cada carga de página real,
// y ListDevices("", filter) legítimamente (y silenciosamente — sin error) encuentra
// cero filas, porque ningún equipo real tiene tenant_id == "".
//
// Cualquier OTRO caller de esta op (una consulta filtrada explícita de una app propia) puede
// seguir pasando un tenant_id real y obtiene un resultado acotado correcto — este bug
// afecta SOLAMENTE la ruta sin argumentos que la UI usa de verdad en cada carga de página.
//
// github.com/veltylabs/staff_manager ya resuelve este problema exacto:
// Deps lleva un TenantID, New lo requiere, y opListStaff retrocede a él
// cuando los argumentos del caller están vacíos. device_manager no tiene ese campo ni
// fallback. Este test siembra un equipo bajo el tenant propio configurado del módulo
// e invoca OpListDevices con un cuerpo VACÍO — exactamente lo que
// callerLister.list() envía — esperando esa fila sembrada de vuelta.
//
// Hoy no compila (devicemanager.Deps no tiene el campo TenantID) —
// ese es el estado rojo esperado; agregar el campo y el fallback es
// la etapa 1 de docs/PLAN.md.
func TestOpListDevices_FallsBackToModuleTenant(t *testing.T) {
	const moduleTenant = "mjosefa-cms"

	db := orm.New(mem.New())
	m, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: moduleTenant})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	seeded, err := m.CreateDevice(devicemanager.Device{
		TenantId: moduleTenant,
		Name:     "Front Desk PC",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}

	reg := &mock.Router{}
	m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	// La forma exacta del cable que view.NewRecords().Caller envía: ningún argumento en absoluto.
	ctx := &mock.Context{InBody: []byte(`{}`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+devicemanager.OpListDevices, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("list_devices with empty args: status = %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
	body := string(ctx.ResponseBody())
	if !contains(body, seeded.Id) {
		t.Fatalf("list_devices with empty args did not return the seeded device (id %q) for the "+
			"module's own tenant %q — got: %s — this reproduces the \"0 / 0\" bug the real UI shows "+
			"on every page load, since it never sends a tenant_id", seeded.Id, moduleTenant, body)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
