---
PLAN: "fix: the tenant always comes from the server (Deps.TenantID); no operation reads tenant_id from the request (D32)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 7076575831483197165
---

> Read AGENTS.md at the repo root before any change.

# PLAN — `device_manager`: tenant from the server only (M2l)

## Why

Phase M2l of `veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md`, decision D32: the tenant never
comes from the browser. The module already has `Deps.TenantID` (`module.go:20`), but several handlers
still prefer what the request sends:

- `ops.go:113` (`opListDevices`): `tenantID := args.TenantId`, falling back to `m.tenantID`.
- `ops.go:138` (`opGetDevice`) and `ops.go:186` (`opDeactivateDevice`): `args.TenantId`, no fallback.
- `ops.go:208` (`opListZones`) and `ops.go:228` (`opListNetworkInterfaces`): `m.tenantOr(args.TenantId)`.
- The three `view.MountSave` functions and `saveDevice` call `m.tenantOr(x.TenantId)`: a record that
  arrives with another tenant is saved under that tenant.
- `opCreateDevice` / `opUpdateDevice` decode a `Device` and keep its `TenantId`.

A person can read or write another installation's devices by sending its tenant. `staff_manager`
v0.1.170 already does it right (its MountSave sets `s.TenantId = m.tenantID`).

The exported Go methods (`GetDevice(tenantID, id)`, `ListZones(tenantID)`, `FindByIP`, …) keep their
`tenantID` parameter: server code and `seed` call them. Only the **handlers** change.

## APIs used

| Symbol | Exact | Where |
|---|---|---|
| `view.MountSave` | already used in `ops.go` | `webtyp.com/view` |
| `loopback.ActingAs` | `func ActingAs(tenantID, userID string, mods ...router.OperationModule) router.Caller` | `webtyp.com/router/loopback` (already used in `tests/screen_save_test.go`) |

## Stage 1 — handlers (`ops.go`)

- Every `args.TenantId` and every `m.tenantOr(...)` in a handler or in a `MountSave` function becomes
  `m.tenantID`.
- Every record received from the request gets `x.TenantId = m.tenantID` **before** the domain call,
  overwriting what came: `saveDevice`, the `OpSaveZone` and `OpSaveNetworkInterface` functions,
  `opCreateDevice`, `opUpdateDevice`.
- Delete `tenantOr` (`ops.go:195-200`): nothing uses it after this.

## Stage 2 — args without `tenant_id` (`model.go`, then `ormc`)

- Remove the `tenant_id` field from `ListDevicesArgsModel`, `GetDeviceArgsModel`,
  `DeactivateDeviceArgsModel`, `ListZonesArgsModel`, `ListNetworkInterfacesArgsModel`.
- `DeleteDeviceArgsModel`, `DeleteZoneArgsModel`, `DeleteNetworkInterfaceArgsModel`: the deletes go
  through `view.MountDelete` since v0.2.85 and nothing decodes them. Run
  `grep -rn "DeleteDeviceArgs\|DeleteZoneArgs\|DeleteNetworkInterfaceArgs" --include=*.go . | grep -v _orm.go`;
  if only `model.go` names them, delete the three definitions.
- **Do not** touch `tenant_id` of the tables (`DeviceModel`, `ZoneModel`, `NetworkInterfaceModel`).
- Regenerate `model_orm.go` with the repo's `ormc` (v0.1.32 or later).

## Stage 3 — tests (`tests/`)

- Tests that send `TenantId` in args stop sending it.
- New `TestScreen_SaveIgnoresTenantFromRequest` in `tests/screen_save_test.go` (same build tag and
  setup as `TestScreenSave` there): through the screen's caller
  (`view.NewRecords().Caller(loopback.ActingAs("tenant-A", …, m))` + `view.NewCallerLister` with the
  Devices presenter's ops), save a device whose `TenantId` is `"other-tenant"`; then
  `m.ListDevices("tenant-A", …)` has it and `m.ListDevices("other-tenant", …)` is empty. Same for one
  zone (`m.ListZones`).
- **Prove it**: before the Stage 1 change this test must fail; say in the PR that you saw it red.

## Stage 4 — docs

`README.md` (Spanish), in the section «Pantallas, errores y permisos»: «El tenant lo pone siempre el
servidor (`Deps.TenantID`): ninguna operación lee `tenant_id` de la petición, y un registro que llega
con otro tenant se guarda en el de esta instalación.»

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission: when the check passes, commit, push and open the PR. Stuck 30
  minutes: open the PR listing what fails.
- `git diff --stat main` lists only: `ops.go`, `model.go`, `model_orm.go`, files under `tests/`,
  `README.md`, this plan. No stray scripts.
- `grep -n "args.TenantId\|tenantOr" ops.go` returns nothing.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `ops.go` | compiles; grep empty |
| 2 | `model.go`, `model_orm.go` | compiles |
| 3 | `tests/*` | suite green; new test seen red first |
| 4 | `README.md` | text present |
