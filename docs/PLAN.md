---
PLAN: "feat: network interfaces, zones and access level; webtyp.com/network host source and importer"
TAG: v0.2.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 12738846217564746307
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> Phase F4 of the network administration master plan (private repo `veltylabs/mjosefa-cms`; you do
> not need it). **Depends on** `webtyp.com/input` (with `input.MAC`, `CanonicalMAC`,
> `IsLocallyAdministeredMAC`) and `webtyp.com/network` v0.1.0. First lines of work:
> `go get webtyp.com/input@latest webtyp.com/network@latest`. If `input.MAC` or `network.HostSource`
> does not exist in what you get, STOP and report — never add a `replace`, never copy code.

# Plan — `github.com/veltylabs/device_manager`: interfaces, zones, access

Read first: [AGENTS.md](../AGENTS.md) (binding module rules; the domain notes at the bottom were
updated for this plan) and [docs/ARCHITECTURE.md](ARCHITECTURE.md) — **the spec**: entities, rules
(IP auto-assignment, MAC rules, which interfaces become hosts, importer behaviour), ops table. This
plan says *where* and *in what order*; ARCHITECTURE says *what*.

## Development rules

- AGENTS.md whitelist/blacklist (now including `webtyp.com/network`; `network/mem` only in tests).
- `webtyp.com/fmt` only (no `errors`/`strings`/`strconv`/stdlib `fmt`), no `map`, no `reflect`.
- Enum literals only in exported constants; no string literals in logic.
- Never hand-edit `model_orm.go`: run `ormc` at the repo root.
- Tests in `tests/`, runner `gotest ./...`. Never export a symbol only for tests.
- **Keep these signatures unchanged** (other modules call them): `New`, `Deps`, `Module.FindByIP`,
  `IPLocator`, `seed.Load(m *Module, tenantID string) (Data, error)`, the existing op names.

## Design gate

**1. Prior art.**
- **NetBox**: `Device` → many `Interface` (each with MAC) → `IPAddress`; IP ranges / prefixes
  allocate "next available IP". Adopted: device/interface split, zone range with lowest-free
  allocation.
- **UniFi / Omada controllers**: clients identified by MAC, "fixed IP" per client, warning about
  private/random MACs. Adopted: MAC as the network identity, randomized MAC rejected.
- **ISC Kea / RouterOS DHCP reservations**: a reservation is MAC → IP. Adopted: an interface with a
  MAC is exactly one reservation on the router.
- Why different: interfaces without MAC are allowed because this inventory also serves login by IP
  (loopback aliases of the server), which no network tool needs.

**2. Novice-name test.** `Zone`, `NetworkInterface`, `Device.Access`, ops `save_zone`,
`save_network_interface`, `NetworkHosts{Devices, TenantID}.Hosts()` — read aloud, each is what it
says. `NetworkInterface` (not `Interface`, a Go keyword-ish word; not `NIC`, an abbreviation).

**3. Complexity ledger.**
```
Concepts the developer must learn   +2 (Zone, NetworkInterface), +1 field (access); −1 field (device.ip)
Files they must touch to do X       registering a device for the router: 0 extra (the app screens)
Lines at the call site              NetworkHosts{Devices: dm, TenantID: t} — 1
Ways to do the same thing           0 — device.ip is DELETED; the interface is the only place an IP lives
```

**4. Where it belongs.** Inventory is this module's concern; applying it to a router is
`network_manager`'s; the boundary types are `webtyp.com/network`'s.

**5. What it deletes.** `Device.Ip` (field, column in the model, `DeviceModel` entry), the IP check
inside `CreateDevice`/`UpdateDevice`, and `Item().Description = d.Ip`. Verify:
`grep -rn "\.Ip\b\|Ip:" --include=*.go . | grep -v model_orm.go` must only show
`NetworkInterface` usages.

## Stage 1 — `model.go`

1. `DeviceModel`: remove the `ip` field. Add, before `is_active`:
   - `{Name: "zone_id", Type: input.Select(), OmitEmpty: true}` — soft reference (no `Ref`): an
     empty value must not become an FK violation. Options are filled by the UI (Stage 5).
   - `{Name: "access", Type: deviceAccess(), NotNull: true}` where `deviceAccess()` is
     `input.Radio(fmt.KeyValue{Key: network.AccessLocalName, Value: "Local network only"}, fmt.KeyValue{Key: network.AccessInternetFilteredName, Value: "Internet (filtered)"}, fmt.KeyValue{Key: network.AccessInternetName, Value: "Internet"})`.
   - `name`: keep `input.Text()` but give the field an explicit whitelist so router comments import:
     `Permitted{Letters: true, Tilde: true, Numbers: true, Spaces: true, Extra: []rune{'.', ',', '(', ')', '-', '_'}, Minimum: 1, Maximum: 255}`.
     A test (Stage 6) checks `"INTERNET PC-GASTRO-NEW"` validates. If it fails because
     `input.Text()`'s own charset still applies, STOP and report (upstream defect) — do not switch kinds.
2. New `ZoneModel` (table `zone`): `id` (PK, OmitEmpty), `tenant_id` (`model.Text()`, NotNull),
   `name` (`input.Text()`, NotNull, same whitelist as device name, `Maximum: 64`),
   `range_start` (`input.IP()`, NotNull), `range_end` (`input.IP()`, NotNull),
   `updated_at` (`BaseInt_FieldInt`, OmitEmpty).
3. New `NetworkInterfaceModel` (table `network_interface`): `id` (PK, OmitEmpty), `tenant_id`
   (NotNull), `device_id` (`input.Select()`, NotNull, `Ref: &DeviceModel`,
   `DB: &model.FieldDB{RefColumn: "id"}`), `label` (`input.Text()`, NotNull, `Minimum: 1, Maximum: 32`),
   `mac` (`input.MAC()`, OmitEmpty), `ip` (`input.IP()`, OmitEmpty), `updated_at`.
4. Transport args: `ListZonesArgsModel` (`tenant_id`), `DeleteZoneArgsModel` (`tenant_id`, `id`),
   `ListNetworkInterfacesArgsModel` (`tenant_id`, `device_id`), `DeleteNetworkInterfaceArgsModel`
   (`tenant_id`, `id`) — base kinds, like the existing args models.
5. Constants and errors (add to the existing blocks):
   ```go
   const (
       ResourceDevice           = "device"
       ResourceZone             = "zone"
       ResourceNetworkInterface = "network_interface"
       ImportedInterfaceLabel   = "imported"
   )
   var (
       ErrMACAlreadyExists = fmt.Err("network interface mac already exists")
       ErrRandomizedMAC    = fmt.Err("randomized MAC address: on the device, turn off random (private) hardware addresses for this Wi-Fi network and register its real MAC")
       ErrNoZone           = fmt.Err("ip is empty and the device has no zone to assign one from")
       ErrIPOutsideZone    = fmt.Err("ip is outside the device's zone range")
       ErrZoneFull         = fmt.Err("zone has no free ip")
       ErrZoneOverlap      = fmt.Err("zone range overlaps another zone")
       ErrZoneInUse        = fmt.Err("zone is used by a device")
       ErrZoneNotFound     = fmt.Err("zone not found")
       ErrInterfaceNotFound = fmt.Err("network interface not found")
   )
   const (
       TopicZoneSaved              = "device_manager.zone.saved"
       TopicZoneDeleted            = "device_manager.zone.deleted"
       TopicNetworkInterfaceSaved   = "device_manager.network_interface.saved"
       TopicNetworkInterfaceDeleted = "device_manager.network_interface.deleted"
   )
   ```
   Use `ResourceDevice` in the existing `Requires("device", …)` calls (replace the literal).
6. Run `ormc`.

## Stage 2 — services

`module.go` (devices):
- `CreateDevice`/`UpdateDevice`: drop all IP code; validate `access` with `network.ParseAccess`
  (error → `ValidationError`); when `zone_id != ""` the zone must exist in the tenant
  (`ErrZoneNotFound` → `ValidationError`); empty `access` on create defaults to
  `network.AccessLocalName` before `Validate`.
- `DeleteDevice`: delete the device's interfaces (tenant-conditioned) first, then the device.
- `FindByIP(tenantId, ip)`: look up `network_interface` by `input.CanonicalIP(ip)` + tenant, then
  return `GetDevice(tenantId, iface.DeviceId)`. Not found → `ErrNotFound` (unchanged contract).

`zone.go` (new): `ListZones(tenantId)`, `SaveZone(z)` (create when `Id == ""`, else update;
validate; both IPs IPv4 and `start <= end` compared numerically — write one unexported
`ipv4ToUint32(ip string) (uint32, bool)` in `ip.go` and reuse it everywhere; overlap with any other
zone of the tenant → `ErrZoneOverlap`; publish `TopicZoneSaved`), `DeleteZone(tenantId, id)`
(`ErrZoneInUse` if any device has that `zone_id`; publish `TopicZoneDeleted`).

`interface.go` (new): `ListNetworkInterfaces(tenantId, deviceId string)` (empty `deviceId` = all),
`SaveNetworkInterface(i)`, `DeleteNetworkInterface(tenantId, id)`. `SaveNetworkInterface`, in order:
1. device must exist in tenant (`ErrNotFound` → 404 at the op);
2. `mac`: if non-empty → `input.CanonicalMAC`; `input.IsLocallyAdministeredMAC` →
   `ValidationError{ErrRandomizedMAC}`; another interface of the tenant with that MAC →
   `ErrMACAlreadyExists`;
3. `ip`: empty → auto-assign (device zone required → `ValidationError{ErrNoZone}`; lowest free IPv4
   in the zone range not used by any interface of the tenant; none → `ErrZoneFull`); non-empty →
   `CanonicalIP`, must be inside the zone when the device has one (`ValidationError{ErrIPOutsideZone}`),
   another interface with that IP → `ErrIPAlreadyExists` (existing sentinel);
4. `Validate(action)`, write, publish `TopicNetworkInterfaceSaved`.

`network.go` (new):
```go
// NetworkHosts adapts the inventory of one tenant to webtyp.com/network.
type NetworkHosts struct {
	Devices  *Module
	TenantID string
}

var (
	_ network.HostSource   = NetworkHosts{}
	_ network.HostImporter = NetworkHosts{}
)

func (n NetworkHosts) Hosts() ([]network.Host, error)
func (n NetworkHosts) ImportHosts(found []network.Discovered) (network.ImportResult, error)
```
Behaviour exactly as ARCHITECTURE "Network boundary". Skip reasons are exported constants:
`SkipNoMAC = "no MAC"`, `SkipNoIP = "no IP"`, `SkipRandomizedMAC = "randomized MAC"`,
`SkipAlreadyRegistered = "MAC already registered"`; a validation failure uses the error text.
Imported device name: `found.Name`, or `"Imported " + MAC` when empty. A Go-side error from the DB
aborts the import and is returned (never swallowed into a skip).

## Stage 3 — `ops.go`

Add the six ops of the ARCHITECTURE table (`save_*` require `model.Create|model.Update`). Handler
shape and status mapping per AGENTS.md; in addition: `ErrMACAlreadyExists`, `ErrZoneOverlap`,
`ErrZoneInUse`, `ErrZoneFull` → 409; `ErrZoneNotFound`, `ErrInterfaceNotFound` → 404;
`ValidationError` → 400. List ops fall back to `m.tenantID` like `opListDevices`.

## Stage 4 — `view.go`

- `Device.Item()`: `Description` = access name + (location if any), no IP.
- `NewZoneView(caller) view.Presenter`, `NewNetworkInterfaceView(caller) view.Presenter` (same
  pattern as `NewView`); `Zone.Item()` → `Label` name, `Description` `"<start> – <end>"`;
  `NetworkInterface.Item()` → `Label` label, `Description` `"<ip> · <mac or 'sin MAC'>"`.

## Stage 5 — `ui/`, `migrate/`, `seed/`, `web/`

- `ui/browser.go`: keep `Browser` (devices). After the crudview is created, call the zones op through
  `caller` and `v.SetOptions("zone_id", <KeyValue{zone.Id, zone.Name}>…)` (see
  `webtyp.com/layout/crudview` `SetOptions`: it repaints a live select). Add
  `ZonesBrowser(caller, ids, tenantID)` (ID `device_manager_zones`, label `Zonas`) and
  `InterfacesBrowser(caller, ids, tenantID)` (ID `device_manager_interfaces`, label
  `Interfaces de red`, `SetOptions("device_id", …)` from `list_devices`). Same signature as `Browser`.
- `migrate/migrate.go`: create tables in FK order: `Zone`, `Device`, `NetworkInterface`. Nothing else
  (see ARCHITECTURE "Existing databases").
- `seed/seed.go`: keep `Load`'s signature and the IP constants. Create one zone `"Recepción"`
  `192.168.1.10`–`192.168.1.29`; the three devices without IP, in that zone, `access=local`
  (the printer) / `internet` (the two computers); one interface each (`label` `ethernet`) with the
  existing IP constants and fixed non-random demo MACs as new constants
  (`ReceptionMAC = "00:1A:2B:00:01:10"`, `Box1MAC = "00:1A:2B:00:01:11"`,
  `PrinterMAC = "00:1A:2B:00:01:20"`). `Data` gains `Zones []devicemanager.Zone` and
  `Interfaces []devicemanager.NetworkInterface`.
- `web/client.go`: register the two new browsers.

## Stage 6 — tests (`tests/`)

Rewrite every existing test that sets `Device.Ip` to create the device and then an interface with
that IP; `FindByIP` / `IPLocator` / tenant tests must still pass with that change only. New files:

- `tests/zone_test.go`: save/list/delete; `start > end` → 400; overlap → `ErrZoneOverlap`; delete in
  use → `ErrZoneInUse`; tenant isolation.
- `tests/interface_test.go`: explicit IP; empty IP auto-assigns the lowest free (`.10`, then `.11`
  after `.10` is taken, skipping a hand-set `.12`); zone full → `ErrZoneFull`; no zone + empty IP →
  `ErrNoZone`; IP outside zone → `ErrIPOutsideZone`; device without zone accepts `127.0.0.1` and
  `::1`; MAC `48-f1-7f-d9-d7-b7` stored as `48:F1:7F:D9:D7:B7`; randomized `4A:C5:93:7A:12:DE` →
  `ErrRandomizedMAC`; duplicate MAC → `ErrMACAlreadyExists`; delete device deletes its interfaces;
  `FindByIP` finds the device through an interface.
- `tests/network_test.go`: `NetworkHosts.Hosts()` returns only active devices' interfaces with MAC,
  with `Name` `"PC 1 (wifi)"` and the device access; an interface without MAC is absent.
  `ImportHosts` with 6 `Discovered` (ok-internet, ok-local, no MAC, no IP, randomized, already
  registered) → `Created == 2`, 4 skipped with the exact reason constants; the imported internet
  device has `access=internet` and the zone containing its IP; `"INTERNET PC-GASTRO-NEW"` imports.
- `tests/conformance_test.go` (existing): extend for the new presenters if it covers views.
- Widget regression: `form.New` over `Device` yields `name`, `type`, `location`, `zone_id`, `access`,
  `is_active` inputs and no `ip`; over `NetworkInterface`: `device_id`, `label`, `mac`, `ip`.

## Stage 7 — docs

`README.md`: ops table from ARCHITECTURE, quick start with `NetworkHosts`, key files (`zone.go`,
`interface.go`, `network.go`, `ip.go`). `docs/diagrams/database.md`: ERD with `zone`, `device`
(`zone_id` dotted soft link), `network_interface` (FK to `device`). Verify ARCHITECTURE against the
code, then **remove the STATUS note** at its top.

## Acceptance criteria

- `gotest ./...` green; `GOOS=js GOARCH=wasm go build .` succeeds.
- `grep -rn "Ip:\|\.Ip\b" --include=*.go . | grep -v model_orm.go | grep -vi interface` → empty.
- `grep -rn '"device"' --include=*.go . | grep -v model_orm.go | grep -v "ResourceDevice\s*=" ` → empty.
- `grep -rn "network/mem" --include=*.go . | grep -v "^./tests/"` → empty.
- `grep -n "STATUS (remove" docs/ARCHITECTURE.md` → empty.

| Stage | Files | Done when |
|---|---|---|
| 1 | `model.go`, `model_orm.go` | ormc generated, no `ip` on Device |
| 2 | `module.go`, `zone.go`, `interface.go`, `network.go`, `ip.go` | compiles |
| 3 | `ops.go` | 13 ops mounted |
| 4 | `view.go` | 3 presenters |
| 5 | `ui/`, `migrate/`, `seed/`, `web/` | demo runs with `webtyp` |
| 6 | `tests/*.go` | green |
| 7 | `README.md`, `docs/diagrams/database.md`, ARCHITECTURE | STATUS note removed |
