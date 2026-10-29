# device_manager
<img src="docs/img/badges.svg">

Módulo de inventario de equipos de red para el ecosistema Velty: registro por tenant de computadores,
impresoras, servidores y otros equipos, sus interfaces de red (MAC + IP), las zonas en las que se ubican
y el nivel de acceso a la red permitido para cada uno. Es la fuente de verdad que `veltylabs/network_manager`
aplica al router a través de `webtyp.com/network`.

## Operaciones

| Operación | Acción | Recurso |
|---|---|---|
| `list_devices` | lectura | `device` |
| `get_device` | lectura | `device` |
| `create_device` | creación | `device` |
| `update_device` | actualización | `device` |
| `upsert_device` | creación/actualización | `device` |
| `deactivate_device` | actualización | `device` |
| `delete_device` | eliminación | `device` |
| `list_zones` | lectura | `zone` |
| `save_zone` | creación/actualización | `zone` |
| `delete_zone` | eliminación | `zone` |
| `list_network_interfaces` | lectura | `network_interface` |
| `save_network_interface` | creación/actualización | `network_interface` |
| `delete_network_interface` | eliminación | `network_interface` |

## Inicio rápido

```go
import devicemanager "github.com/veltylabs/device_manager"

dm, err := devicemanager.New(db, devicemanager.Deps{
    IDs:       idGenerator,    // model.IDGenerator, requerido
    Publisher: eventPublisher, // events.Publisher, opcional
})
dm.MountOperations(reg)    // router.OperationRegistry
view := dm.NewView(caller) // router.Caller -> view.Presenter (también NewZoneView, NewNetworkInterfaceView)

// Frontera con webtyp.com/network: los hosts que el router debe aplicar, y la
// importación inicial de configuraciones manuales del router.
hosts := devicemanager.NetworkHosts{Devices: dm, TenantID: tenant} // network.HostSource + network.HostImporter
```

## Vistas y demo

El módulo exporta tres pantallas declarativas bajo `ui/` accesibles mediante `ui.Screens(caller)` (Equipos, Zonas, e Interfaces de red).
Cada pantalla cuenta con su propio dibujo tomado de `webtyp/icons` y una descripción (`view.WithDescription`). Los selectores
de zonas y equipos se configuran declarativamente mediante `Lookups`.

Los datos de demostración pueden cargarse usando `seed.Load(m, tenantID)`.

## Archivos clave

| Archivo | Propósito |
|---|---|
| `model.go` | Definiciones de `Device`, `Zone`, `NetworkInterface`, errores y tópicos de eventos |
| `kinds/` | Tipos cerrados para campos con opciones (tipo de equipo, acceso) — paquete propio para ormc |
| `model_orm.go` | Generado por `ormc` — no editar |
| `module.go` | `Module`, `New`, métodos de servicio para equipos |
| `zone.go`, `interface.go`, `ip.go` | Zonas, interfaces (reglas MAC/IP, autoasignación de IP), utilidades IPv4 |
| `network.go` | `NetworkHosts`: implementación de `HostSource` y `HostImporter` de `webtyp.com/network` |
| `ops.go` | `router.OperationModule` — `MountOperations` y manejadores |
| `view.go` | `NewView`, `NewZoneView`, `NewNetworkInterfaceView` |

## Documentación

| Documento | Descripción |
|---|---|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Alcance del dominio, entidades, patrones, tabla de operaciones |
| [Diagrama de base de datos](docs/diagrams/database.md) | Diagrama ER en Mermaid |

## Pantallas, errores y permisos

Las pantallas guardan y borran con `view.MountSave` / `view.MountDelete`; los errores salen por
`rpc.WriteError` (cada error del módulo es `model.NotFound`, `model.Conflict` o `model.Invalid`), y los
permisos se nombran con la tabla: `model.ResourceOf(&devicemanager.DeviceModel)`.
