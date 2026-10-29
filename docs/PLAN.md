---
PLAN: "refactor: screens save and delete through view.MountSave/MountDelete; errors through rpc.WriteError; resources from the tables; English menu keys"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 9389973514273114292
---

> Read AGENTS.md at the repo root before any change.

# PLAN — `device_manager` passes the module review (M2d)

## Why

The Devices, Zones and Network interfaces screens **cannot save**: the generic screen sends
`{"records":[…]}` and deletes with `{"ids":[…]}`, but `opUpsertDevice`, `opSaveZone`,
`opSaveNetworkInterface` decode a single record, and the deletes decode `{tenant_id, id}`. Errors
answer 400/404/409 with an empty body. Phase M2d of
`veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md` (decisions D22, D26, D27). The framework now
has the pieces: `view.MountSave`/`view.MountDelete` (view v0.6.53) read exactly what the screen
sends; `rpc.WriteError` (rpc v0.2.10) answers every error; `model.ResourceOf` (model v0.2.12) names a
table's permission; the error kinds are `model.NotFound`/`Conflict`/`Invalid`.

## APIs used

| Symbol | Exact | Tag |
|---|---|---|
| `view.MountSave` | `func MountSave[T any, PT interface{ *T; model.Model }](reg router.OperationRegistry, op string, save func(PT) error) router.Route` | view v0.6.53 |
| `view.MountDelete` | `func MountDelete(reg router.OperationRegistry, op string, remove func(ids []string) error) router.Route` | view v0.6.53 |
| `rpc.WriteError` | `func WriteError(ctx router.Context, err error)` | rpc v0.2.10 |
| `model.ResourceOf` | `func ResourceOf(def *Definition) Resource` | model v0.2.12 |
| `model.NotFound`, `model.Conflict`, `model.Invalid` | `type X string` with `Error()` and `Kind()`; `const ErrX model.NotFound = "…"` | model v0.2.12 |
| `model.FieldError` | `struct{ Field string; Err error }` | model v0.2.12 |
| `model.KindOf` | `func KindOf(err error) ErrorKind` | model v0.2.12 |

## Tools

```
go get webtyp.com/view@v0.6.53 webtyp.com/rpc@v0.2.10 webtyp.com/model@v0.2.12
```

## Stage 1 — errors (`model.go`)

Delete `domainError`, `ValidationError`. Re-type every `Err*` constant keeping its text, so today's
statuses do not change:

| Constants | Type |
|---|---|
| `ErrNotFound`, `ErrZoneNotFound`, `ErrInterfaceNotFound` | `model.NotFound` |
| `ErrIPAlreadyExists`, `ErrMACAlreadyExists`, `ErrZoneOverlap`, `ErrZoneInUse` | `model.Conflict` |
| `ErrRandomizedMAC`, `ErrNoZone`, `ErrIPOutsideZone`, `ErrZoneFull`, `ErrInvalidZoneRange` | `model.Invalid` |

Add `ErrCannotRead model.Invalid = "device_manager: cannot read the request"`.
`ValidationError{Err: fmt.Err("invalid device type: %s", d.Type)}` (`module.go:134`) becomes
`&model.FieldError{Field: "type", Err: fmt.Err("invalid device type: %s", d.Type)}`; any other
`ValidationError{Err: err}` returns `err` (already a `*model.FieldError`). Every
`err.(domainError); ok && e == ErrX` and `err == ErrX` becomes `model.KindOf(err) == model.Error…`
(no `==` between errors anywhere: it pulls reflection into the browser binary).

## Stage 2 — resources (`model.go`)

Delete `ResourceDevice`, `ResourceZone`, `ResourceNetworkInterface`. Every use becomes
`model.ResourceOf(&DeviceModel)`, `model.ResourceOf(&ZoneModel)`,
`model.ResourceOf(&NetworkInterfaceModel)` — in `ops.go` and in `ui/browser.go` (`Permission:`).
The names do not change (`device`, `zone`, `network_interface`).

## Stage 3 — save and delete (`ops.go`)

- `OpUpsertDevice`: `view.MountSave(reg, OpUpsertDevice, func(d *Device) error { out, err := m.saveDevice(*d); *d = out; return err }).Requires(model.ResourceOf(&DeviceModel), model.Create|model.Update)`,
  where `saveDevice` is the create-or-update logic `opUpsertDevice` has today (id empty → create;
  else update), moved to an unexported method.
- `OpSaveZone`: same with `m.SaveZone`. `OpSaveNetworkInterface`: same with `m.SaveNetworkInterface`.
- `OpDeleteDevice`, `OpDeleteZone`, `OpDeleteNetworkInterface`: `view.MountDelete(reg, op, func(ids []string) error { for each id: m.DeleteX(m.tenantID, id); first error returned })`.
- Delete `opUpsertDevice`, `opSaveZone`, `opSaveNetworkInterface`, the three `opDelete*` handlers,
  and `DeleteDeviceArgs`/`DeleteZoneArgs`/`DeleteNetworkInterfaceArgs` (and their models in
  `model.go`; regenerate `model_orm.go` with `ormc`) if nothing else uses them.
- Every other handler (`list`, `get`, `create`, `update`, `deactivate`) writes errors with
  `rpc.WriteError(ctx, err)`; a decode failure is `rpc.WriteError(ctx, ErrCannotRead)`. No
  `ctx.WriteStatus(4xx/5xx)` left.

## Stage 4 — menu keys (`ui/module.go`, `lang.json`)

`Label = "Devices"`, `ZonesLabel = "Zones"`, `InterfacesLabel = "Network interfaces"` (English
keys). Add them to `lang.json` translated to every language the file has (es: «Equipos», «Zonas»,
«Interfaces de red»); `langc sync .` → nothing pending, no reuse warnings.

## Stage 5 — tests

- Update tests that used the deleted types or handlers.
- New `tests/screen_save_test.go`: through `view.NewRecords().Caller(loopback.ActingAs(...))` and
  `view.NewCallerLister` with each presenter's ops (as an app does), save a device, a zone and a
  network interface, and delete each: no error, and the store has / no longer has the row. Plus:
  saving a device with a duplicate IP answers `model.KindOf(err) == model.ErrorConflict` on the
  client.
- Demo (`web/browser/main.go` if present): open it with `webtyp`, save one device; screenshot in the
  PR description.

## Stage 6 — docs (Spanish)

`README.md`: las pantallas guardan y borran con `view.MountSave`/`MountDelete`; errores por
`rpc.WriteError`; permisos por tabla con `model.ResourceOf`.

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission: when the check passes, commit, push and open the PR. Stuck 30
  minutes: open the PR listing what fails.
- `git diff --stat main` lists only files of this repo touched by the stages, `go.mod`, `go.sum`,
  this plan. No stray scripts.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `model.go`, `module.go`, `zone.go`, `interface.go`, `network.go` | compiles |
| 2 | `model.go`, `ops.go`, `ui/browser.go` | compiles |
| 3 | `ops.go`, `model.go`, `model_orm.go` | compiles |
| 4 | `ui/module.go`, `lang.json` | `langc sync .` clean |
| 5 | `tests/*` | suite green |
| 6 | `README.md` | text present |
