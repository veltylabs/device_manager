---
PLAN: "Refactor UI package to use declarative `[]view.Screen` and remove `webtyp.com/layout` dependency."
TAG: ui-refactor
EXECUTOR: Jules
REVIEWER: auto
---

## Executor notes
The `webtyp.com/lang` module reported missing translations for `"is required"`. After extracting it from `"device_manager: Deps.IDs is required"` to prevent duplication, I updated the corresponding references in `module.go` with `lang.Text("is required")`. The translation logic was verified with `go run webtyp.com/lang/cmd/langc@latest sync .`

## Stage 1: Actualizar constructores de Presenter

En `view.go`:
Agregar el soporte de opciones variables `opts ...view.Option` a los tres constructores de vistas.

```go
package devicemanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

// Item implementa view.Itemizer para representar un equipo en listas.
//
// Ejemplo: item := d.Item()
func (d *Device) Item() view.Item {
	desc := d.Access
	if d.Location != "" {
		desc += " · " + d.Location
	}
	return view.Item{ID: d.Id, Label: d.Name, Description: desc}
}

// Item implementa view.Itemizer para representar una zona en listas.
//
// Ejemplo: item := z.Item()
func (z *Zone) Item() view.Item {
	return view.Item{ID: z.Id, Label: z.Name, Description: z.RangeStart + " – " + z.RangeEnd}
}

// noMACLabel describe una interfaz que el router no gestiona (sin MAC).
const noMACLabel = "sin MAC"

// Item implementa view.Itemizer para representar una interfaz de red en listas.
//
// Ejemplo: item := ni.Item()
func (ni *NetworkInterface) Item() view.Item {
	mac := ni.Mac
	if mac == "" {
		mac = noMACLabel
	}
	return view.Item{ID: ni.Id, Label: ni.Label, Description: ni.Ip + " · " + mac}
}

const titleDevices = "Equipos"

// NewView construye el Presenter de equipos — el motor independiente de la tecnología
// que cualquier renderizador (crudview u otro) envuelve.
//
// Ejemplo: p := devicemanager.NewView(caller, view.WithDescription("Ayuda..."))
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListDevices, Save: OpUpsertDevice, Delete: OpDeleteDevice},
		func() model.ModelSlice { return &DeviceList{} })
	allOpts := append([]view.Option{view.WithTitle(titleDevices)}, opts...)
	return view.New(b, &Device{}, allOpts...)
}

const (
	titleZones             = "Zonas"
	titleNetworkInterfaces = "Interfaces de red"
)

// NewZoneView construye el Presenter de zonas.
//
// Ejemplo: p := devicemanager.NewZoneView(caller, view.WithDescription("Ayuda..."))
func NewZoneView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListZones, Save: OpSaveZone, Delete: OpDeleteZone},
		func() model.ModelSlice { return &ZoneList{} })
	allOpts := append([]view.Option{view.WithTitle(titleZones)}, opts...)
	return view.New(b, &Zone{}, allOpts...)
}

// NewNetworkInterfaceView construye el Presenter de interfaces de red.
//
// Ejemplo: p := devicemanager.NewNetworkInterfaceView(caller, view.WithDescription("Ayuda..."))
func NewNetworkInterfaceView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkInterfaces, Save: OpSaveNetworkInterface, Delete: OpDeleteNetworkInterface},
		func() model.ModelSlice { return &NetworkInterfaceList{} })
	allOpts := append([]view.Option{view.WithTitle(titleNetworkInterfaces)}, opts...)
	return view.New(b, &NetworkInterface{}, allOpts...)
}
```

---

## Stage 2: Declarar ui.Screens sin dependencias de layout

En `ui/browser.go`:
Eliminar los métodos `Browser`, `ZonesBrowser`, `InterfacesBrowser`, `onFirstActivation` y `fillOptions`.
Reemplazar con `Screens(caller router.Caller) []view.Screen`.

Reescribir `ui/browser.go`:

```go
package ui

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

// Campos select cuyas opciones provienen de otra lista de este módulo.
const (
	fieldZoneID   = "zone_id"
	fieldDeviceID = "device_id"
)

// Textos de ayuda que el chasis muestra en el botón «?» de cada pantalla. Las
// claves están en inglés; la traducción vive en lang.json.
const (
	descDevices    lang.Text = "Each computer, printer or device of the clinic, with the zone where it is. Its network interfaces are added in Network interfaces."
	descZones      lang.Text = "Zones group the devices by place (for example, reception or box 1). Create the zones first; each device then picks one."
	descInterfaces lang.Text = "The network connections of each device: its MAC and IP address. Staff log in from the devices registered here."
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
//
// Ejemplo: screens := ui.Screens(caller)
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:         ID,
			Label:      Label,
			Icon:       laptop.Name,
			Permission: devicemanager.ResourceDevice,
			Presenter:  devicemanager.NewView(caller, view.WithDescription(descDevices)),
			Lookups: []view.Lookup{
				{Field: fieldZoneID, Source: devicemanager.NewZoneView(caller)},
			},
		},
		{
			ID:         ZonesID,
			Label:      ZonesLabel,
			Icon:       location.Name,
			Permission: devicemanager.ResourceZone,
			Presenter:  devicemanager.NewZoneView(caller, view.WithDescription(descZones)),
		},
		{
			ID:         InterfacesID,
			Label:      InterfacesLabel,
			Icon:       network.Name,
			Permission: devicemanager.ResourceNetworkInterface,
			Presenter:  devicemanager.NewNetworkInterfaceView(caller, view.WithDescription(descInterfaces)),
			Lookups: []view.Lookup{
				{Field: fieldDeviceID, Source: devicemanager.NewView(caller)},
			},
		},
	}
}
```

---

## Stage 3: Eliminar directorio web/

Eliminar el archivo `web/browser/main.go` y la carpeta `web/`:
```bash
rm -rf web
```

---

## Stage 4: Actualizar pruebas unitarias

1. En `tests/ui_described_test.go`:
Reescribir para validar `ui.Screens` y asegurar que no importa `webtyp.com/layout`:

```go
package tests

import (
	"testing"

	devicemanager "github.com/veltylabs/device_manager"
	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/view"
)

// Las tres pantallas se explican solas, validan según view.Screen y cada una tiene su propio dibujo y permiso.
func TestUI_ScreensAreDescribedWithTheirOwnIcon(t *testing.T) {
	fake := &uiFakeCaller{}
	screens := ui.Screens(fake)
	if len(screens) != 3 {
		t.Fatalf("expected 3 screens, got %d", len(screens))
	}

	want := []struct {
		id         string
		icon       string
		permission string
	}{
		{ui.ID, laptop.Name, string(devicemanager.ResourceDevice)},
		{ui.ZonesID, location.Name, string(devicemanager.ResourceZone)},
		{ui.InterfacesID, network.Name, string(devicemanager.ResourceNetworkInterface)},
	}

	for i, w := range want {
		s := screens[i]
		if err := s.Validate(); err != nil {
			t.Fatalf("screen %s validation failed: %v", s.ID, err)
		}
		if s.ID != w.id {
			t.Errorf("screen %d ID = %q, want %q", i, s.ID, w.id)
		}
		if s.Icon != w.icon {
			t.Errorf("screen %d Icon = %q, want %q", i, s.Icon, w.icon)
		}
		if string(s.Permission) != w.permission {
			t.Errorf("screen %d Permission = %q, want %q", i, s.Permission, w.permission)
		}
		d, ok := s.Presenter.(view.WithDescription)
		if !ok || d.Description() == "" {
			t.Errorf("screen %d presenter must implement view.WithDescription with a non-empty text", i)
		}
	}
}
```

2. En `tests/ui_view_test.go`:
Actualizar para probar que `ui.Screens` declara los select lookups correspondientes:

```go
package tests

import (
	"testing"

	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/model"
)

type uiFakeCaller struct {
	calls []string
}

func (f *uiFakeCaller) Call(op string, args model.Encodable, out model.Decodable, done func(error)) {
	f.calls = append(f.calls, op)
	if done != nil {
		done(nil)
	}
}

func (f *uiFakeCaller) Dispatch(op string, args model.Encodable) {}

func TestUIView_Lookups(t *testing.T) {
	fake := &uiFakeCaller{}
	screens := ui.Screens(fake)

	// Dispositivos debe tener lookup en zone_id
	devScreen := screens[0]
	if len(devScreen.Lookups) != 1 {
		t.Fatalf("expected 1 lookup on devices screen, got %d", len(devScreen.Lookups))
	}
	if devScreen.Lookups[0].Field != "zone_id" {
		t.Errorf("lookup field = %q, want zone_id", devScreen.Lookups[0].Field)
	}
	if devScreen.Lookups[0].Source == nil {
		t.Errorf("lookup source must not be nil")
	}

	// Interfaces debe tener lookup en device_id
	ifaceScreen := screens[2]
	if len(ifaceScreen.Lookups) != 1 {
		t.Fatalf("expected 1 lookup on interfaces screen, got %d", len(ifaceScreen.Lookups))
	}
	if ifaceScreen.Lookups[0].Field != "device_id" {
		t.Errorf("lookup field = %q, want device_id", ifaceScreen.Lookups[0].Field)
	}
	if ifaceScreen.Lookups[0].Source == nil {
		t.Errorf("lookup source must not be nil")
	}
}
```

---

## Stage 5: Limpieza de dependencias en go.mod y verificación

Ejecutar:
```bash
go mod tidy
```

Verificar que:
1. `grep "webtyp.com/layout" go.mod` no devuelve nada.
2. `grep "webtyp.com/layout" go.sum` no devuelve nada.
3. `grep -rn "webtyp.com/layout" .` no devuelve nada.

Ejecutar la suite de pruebas completa:
```bash
gotest
```

---

## Stage 6: Documentación y regla de idioma en veltylabs

1. En `AGENTS.md`:
Actualizar la sección `ui/, seed/ and web/`:
Reemplazar la referencia a `Browser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error)` por:
`Screens(caller router.Caller) []view.Screen`.
Indicar claramente que `ui/` no importa `webtyp.com/layout` ni ningún renderizador, sino que devuelve datos declarativos (`[]view.Screen`).

2. En `README.md`:
Actualizar cualquier mención a `ui.Browser` por `ui.Screens`.

3. Ejecutar `langc sync .` y verificar que no haya claves vacías ni errores.

---

## Stages table

| Etapa | Archivos modificados/creados | Qué hace |
|---|---|---|
| Stage 1 | `view.go` | Agrega `opts ...view.Option` a constructores de Presenter |
| Stage 2 | `ui/browser.go` | Exporta `ui.Screens` declarativo con `view.Screen` y borra imports de `layout` |
| Stage 3 | `web/` (borrado) | Elimina `web/browser/main.go` |
| Stage 4 | `tests/ui_described_test.go`, `tests/ui_view_test.go` | Valida `ui.Screens`, `Lookups` y remueve dependencias de `layout` |
| Stage 5 | `go.mod`, `go.sum` | `go mod tidy` y ejecución de `gotest` |
| Stage 6 | `AGENTS.md`, `README.md`, `lang.json` | Actualiza documentación en español y sincroniza `langc` |

## Delivery check

1. `grep -rn "webtyp.com/layout" .` debe retornar vacío.
2. `test ! -d web` debe retornar éxito (0).
3. `go test -v ./...` pasa con 0 fallos de compilación, 0 fallos de tests y cobertura verificada.
4. `go run webtyp.com/lang/cmd/langc@latest sync .` pasa sin errores.
5. El PR se abre inmediatamente al pasar estas comprobaciones.
