# Device Manager Architecture

## Domain Scope

`device_manager` owns the **inventory of a site's equipment** (computers, printers, servers, other):
what each device is, where it is (zone), which network cards it has (MAC + IP) and how much of the
network it may use (access level). It is the *source of truth* that `veltylabs/network_manager`
pushes to the router through `webtyp.com/network`; it never talks to a router itself.

It also answers "which device has this IP?" (`FindByIP` / `IPLocator`), which `staff_manager` and the
apps use for login from the local network.

## Entities

- **Zone**: a named area of the site served by one access point, with an IPv4 range
  (`range_start`–`range_end`), e.g. "AP-06 piso 2", `172.0.0.60`–`172.0.0.69`. Ranges of one tenant
  never overlap. A zone in use by a device cannot be deleted.
- **Device**: `name`, `type` (`DeviceTypeComputer`/`Printer`/`Server`/`Other`), `location` (free
  text), `zone_id` (optional, soft reference — empty means "no zone"), `access` (one of the
  `webtyp.com/network` access names: `local`, `internet_filtered`, `internet`; default `local`),
  `is_active`.
- **NetworkInterface**: one network card of a device — `device_id` (FK), `label` ("wifi",
  "ethernet"), `mac` (optional), `ip`.
  - **IP**: unique per tenant. Left empty, it is **assigned automatically**: the lowest free address
    of the device's zone. Given explicitly, it must lie inside the device's zone (devices without a
    zone may hold any IP — e.g. the server's loopback aliases used for login).
  - **MAC**: optional, unique per tenant when present, stored in canonical form
    (`input.CanonicalMAC`). A **randomized** MAC (`input.IsLocallyAdministeredMAC`) is rejected with
    instructions: phones and Windows can change it, and the device would lose its access without
    warning (incident 2026-10-07).
  - An interface **without MAC** still identifies its device for login by IP, but is **not** sent to
    the router: only interfaces with a MAC become `network.Host`s.

Closed-option kinds (`type`, `access` radios) live in the `kinds` subpackage: ormc resolves a field
kind by compiling its constructor, which it only imports from a package of its own. The root package
re-exports the type names (`DeviceTypeComputer` = `kinds.TypeComputer`, …).

## Network boundary (`webtyp.com/network`)

- `NetworkHosts{Devices, TenantID}` implements `network.HostSource`: one `network.Host` per interface
  with a MAC of an **active** device — `Name` = `"<device name> (<label>)"`, `Access` = the device's.
- The same type implements `network.HostImporter`: creates a device (type `other`, access `internet`
  if the router granted it Internet by hand, else `local`, zone = the zone whose range contains the
  IP) plus one interface (`label` = `imported`) per discovered MAC. An entry without a router comment
is named `Imported <MAC with '-' separators>` (`:` is not allowed in a name). Skipped, with a reason: no MAC,
  no IP, randomized MAC, MAC already registered, or a name that fails validation.

## Patterns

- **Reusable-module harness**: coupled only to published contracts — see `AGENTS.md` (repo root).
  `orm.DB` for storage; `migrate/` for schema; `router.OperationModule` for transport;
  `model.IDGenerator` for identity; optional `events.Publisher`; `view.Presenter` for UI;
  `webtyp.com/network` for the router boundary.
- **Multi-tenancy**: every row carries `tenant_id`; every read/update/delete condition includes it.
- **Typed events**: every published event carries the typed record.

## Ops (via `MountOperations`)

| Op | Action | Resource | Description |
|---|---|---|---|
| `list_devices` | `r` | `device` | Devices of a tenant, optional type/active filter |
| `get_device` | `r` | `device` | One device |
| `create_device` / `update_device` / `upsert_device` | `c` / `u` / `c`+`u` | `device` | Write a device |
| `deactivate_device` | `u` | `device` | Soft-delete (`is_active = false`) — its hosts leave the router plan |
| `delete_device` | `d` | `device` | Hard-delete, with its interfaces |
| `list_zones` | `r` | `zone` | Zones of a tenant |
| `save_zone` | `c`+`u` | `zone` | Create or update a zone |
| `delete_zone` | `d` | `zone` | Delete an unused zone |
| `list_network_interfaces` | `r` | `network_interface` | Interfaces, optionally of one device |
| `save_network_interface` | `c`+`u` | `network_interface` | Create or update; empty IP is auto-assigned |
| `delete_network_interface` | `d` | `network_interface` | Delete an interface |

## Existing databases

Databases created before this model have a `device.ip` column (NOT NULL). `migrate.Migrate` creates
`zone` and `network_interface`; copying each old `device.ip` into an interface and dropping the old
column is a one-time step of the consuming application (`webtyp.com/ddl` never drops a populated
column on its own).

## Composition Root Example

```go
dm, _ := devicemanager.New(db, devicemanager.Deps{IDs: ids, Publisher: pub, TenantID: tenant})
dm.MountOperations(reg)
hosts := devicemanager.NetworkHosts{Devices: dm, TenantID: tenant} // network.HostSource + HostImporter
```
