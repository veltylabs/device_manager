# Arquitectura de Device Manager

## Alcance del dominio

`device_manager` gestiona el **inventario de equipos del recinto** (computadores, impresoras, servidores, otros):
qué es cada equipo, dónde está (zona), qué tarjetas de red posee (MAC + IP) y qué nivel de red puede
usar (nivel de acceso). Es la *fuente de verdad* que `veltylabs/network_manager` aplica al router
a través de `webtyp.com/network`; nunca interactúa directamente con el router.

También responde a "¿qué equipo tiene esta IP?" (`FindByIP` / `IPLocator`), utilizado por `staff_manager`
y las aplicaciones para el inicio de sesión desde la red local.

## Entidades

- **Zone**: área identificada del recinto atendida por un punto de acceso, con un rango IPv4
  (`range_start`–`range_end`), por ejemplo "AP-06 piso 2", `172.0.0.60`–`172.0.0.69`. Los rangos de un
  mismo tenant nunca se solapan. Una zona en uso por un equipo no puede eliminarse.
- **Device**: `name`, `type` (`DeviceTypeComputer`/`Printer`/`Server`/`Other`), `location` (texto libre),
  `zone_id` (referencia blanda opcional; vacía significa "sin zona"), `access` (uno de los nombres de
  acceso de `webtyp.com/network`: `local`, `internet_filtered`, `internet`; por defecto `local`),
  `is_active`.
- **NetworkInterface**: una tarjeta de red de un equipo — `device_id` (clave foránea), `label` ("wifi",
  "ethernet"), `mac` (opcional), `ip`.
  - **IP**: única por tenant. Si se deja vacía, se **asigna automáticamente**: la dirección libre más
    baja de la zona del equipo. Si se especifica explícitamente, debe pertenecer al rango de la zona del
    equipo (los equipos sin zona pueden tener cualquier IP, como las interfaces loopback del servidor usadas
    para inicio de sesión).
  - **MAC**: opcional, única por tenant cuando está presente, almacenada en forma canónica
    (`input.CanonicalMAC`). Las MAC **aleatorias** (`input.IsLocallyAdministeredMAC`) se rechazan con
    instrucciones: teléfonos y Windows pueden cambiarlas y el equipo perdería acceso sin previo aviso.
  - Una interfaz **sin MAC** identifica al equipo para inicio de sesión por IP, pero **no** se envía
    al router: solo las interfaces con MAC se convierten en `network.Host`.

Los tipos cerrados (`type`, radios de `access`) residen en el subpaquete `kinds`: ormc resuelve el tipo
compilando su constructor, el cual solo importa desde un paquete propio. El paquete raíz reexporta los
nombres (`DeviceTypeComputer` = `kinds.TypeComputer`, …).

## Frontera de red (`webtyp.com/network`)

- `NetworkHosts{Devices, TenantID}` implementa `network.HostSource`: un `network.Host` por cada interfaz
  con MAC de un equipo **activo** — `Name` = `"<nombre de equipo> (<etiqueta>)"`, `Access` = el del equipo.
- El mismo tipo implementa `network.HostImporter`: crea un equipo (tipo `other`, acceso `internet` si
  el router le otorgó Internet manualmente, de lo contrario `local`, zona = la zona cuyo rango contiene la
  IP) más una interfaz (`label` = `imported`) por cada MAC descubierta. Una entrada sin comentario del router
  se nombra `Imported <MAC con guiones>` (`:` no está permitido en un nombre). Se omiten con motivo:
  sin MAC, sin IP, MAC aleatoria, MAC ya registrada o un nombre que no valida.

## Patrones

- **Estructura de módulo reutilizable**: acoplado solo a contratos publicados — ver `AGENTS.md` (raíz del repo).
  `orm.DB` para persistencia; `migrate/` para esquema; `router.OperationModule` para transporte;
  `model.IDGenerator` para identidad; opcional `events.Publisher`; `view.Presenter` para interfaz;
  `webtyp.com/network` para la frontera con el router.
- **Multiarrendamiento (Multi-tenancy)**: cada registro lleva `tenant_id`; cada condición de lectura/actualización/eliminación lo incluye.
- **Eventos tipados**: cada evento publicado transporta el registro tipado correspondiente.

## Operaciones (vía `MountOperations`)

| Operación | Acción | Recurso | Descripción |
|---|---|---|---|
| `list_devices` | `r` | `device` | Equipos de un tenant, con filtro opcional por tipo/activo |
| `get_device` | `r` | `device` | Un equipo |
| `create_device` / `update_device` / `upsert_device` | `c` / `u` / `c`+`u` | `device` | Escribir un equipo |
| `deactivate_device` | `u` | `device` | Baja lógica (`is_active = false`) — sus hosts salen del plan del router |
| `delete_device` | `d` | `device` | Eliminación definitiva, junto a sus interfaces |
| `list_zones` | `r` | `zone` | Zonas de un tenant |
| `save_zone` | `c`+`u` | `zone` | Crear o actualizar una zona |
| `delete_zone` | `d` | `zone` | Eliminar una zona en desuso |
| `list_network_interfaces` | `r` | `network_interface` | Interfaces, opcionalmente de un equipo |
| `save_network_interface` | `c`+`u` | `network_interface` | Crear o actualizar; IP vacía se autoasigna |
| `delete_network_interface` | `d` | `network_interface` | Eliminar una interfaz |

## Bases de datos existentes

Las bases de datos creadas antes de este modelo poseen una columna `device.ip` (NOT NULL). `migrate.Migrate`
crea `zone` y `network_interface`; copiar cada `device.ip` antiguo a una interfaz y descartar la columna antigua
es un paso único de la aplicación consumidora (`webtyp.com/ddl` nunca descarta columnas con datos por sí sola).

## Ejemplo de composición

```go
dm, _ := devicemanager.New(db, devicemanager.Deps{IDs: ids, Publisher: pub, TenantID: tenant})
dm.MountOperations(reg)
hosts := devicemanager.NetworkHosts{Devices: dm, TenantID: tenant} // network.HostSource + HostImporter
```
