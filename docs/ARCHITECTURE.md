# Device Manager Architecture

## Domain Scope
`device_manager` owns the registry of network/office equipment (computers, printers, servers, and
other devices) per tenant/location. It is the productionized form of the
[`tinywasm/layout/platformd/modules/devices`](https://github.com/tinywasm/layout/blob/main/platformd/modules/devices/devices.go)
UI demo (`id`/`name`/`ip`), extended with `type`, `location`, and `is_active` for real inventory
management.

## Entities
- **Device**: one piece of equipment — `name`, `ip` (validated via `input.IP()`), `type` (one of
  `DeviceTypeComputer`/`DeviceTypePrinter`/`DeviceTypeServer`/`DeviceTypeOther`), `location`
  (optional), `is_active`. IP uniqueness is enforced **per tenant** in the service layer, not as a
  DB constraint — two tenants may register the same private IP range independently.

## Patterns
- **Reusable-module harness**: coupled only to published contracts — see `AGENTS.md` (this repo's
  root) for the full whitelist/blacklist:
  - `orm.DB` for storage (backend-agnostic); `ddl.CreateTable` (over `db.RawConn()`) for the
    module's own schema migration in `New()`.
  - `router.OpModule` (`ModelName()` + `MountOps(reg router.OpRegistry)`) for transport — the
    module never implements `router.APIModule`/`Router`, never imports `tinywasm/mcp`.
  - `model.IDGenerator` for identity (`Deps.IDs`, required — the module never builds its own).
  - `events.Publisher` for domain events (`Deps.Publisher`, optional — `nil` disables publishing
    silently).
  - `view.Presenter` (`NewView(caller router.Caller) view.Presenter`) for UI, built with only
    `view`+`model`+`router` — the app picks the renderer (`tinywasm/layout/crudview` or any other).
- **Multi-tenancy**: every `Device` row carries `tenant_id`; every read/update/delete condition
  includes it (`orm.Eq(Device_.Id, id), orm.Eq(Device_.TenantId, tenantId)`).
- **Typed events**: every published event carries `*Device` (`model.Encodable`), never a bare map.

## Ops (via `MountOps`)
| Op | Action | Resource | Description |
|---|---|---|---|
| `list_devices` | `r` | `device` | List devices for a tenant, optional type/active filter |
| `get_device` | `r` | `device` | Get device by ID |
| `create_device` | `c` | `device` | Create a new device |
| `update_device` | `u` | `device` | Update an existing device |
| `upsert_device` | `c`/`u` | `device` | Create or update depending on whether `id` is set |
| `deactivate_device` | `u` | `device` | Soft-delete (`is_active = false`) |
| `delete_device` | `d` | `device` | Hard-delete |

## Composition Root Example
```go
dm, _ := devicemanager.New(db, devicemanager.Deps{
    IDs:       idGenerator,    // model.IDGenerator
    Publisher: eventPublisher, // events.Publisher, nil disables publishing
})
dm.MountOps(opRegistry)    // router.OpRegistry
view := dm.NewView(caller) // router.Caller -> view.Presenter
```
