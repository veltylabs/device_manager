---
PLAN: "fix: opListDevices has no tenant fallback, so every real list load (which sends no args) returns zero rows"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 9257265402213424410
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# PLAN — `device_manager`: add `Deps.TenantID` and fall back to it in `opListDevices`

You are an external agent with **zero prior context** about this project. Everything you need is
in this file. Read `AGENTS.md` at the repo root first, then this file fully before writing code.

## 0. Prerequisite — run this first

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

All tests run with `gotest`, never `go test` directly.

## 1. The bug, proven by a failing test already in this repo

`tests/tenant_fallback_test.go` (already committed on this branch) reproduces a real bug found
while manually testing a downstream app's "Equipos" screen: the list showed "0 / 0" even though a
real device existed for the app's tenant. Run it now and confirm it is RED — it does not even
compile yet, which is the expected starting point:

```bash
gotest
# tests/tenant_fallback_test.go:44:72: unknown field TenantID in struct literal of type devicemanager.Deps
```

The acceptance criterion: **`gotest` must go fully green**, including this test, with no other
test regressing.

## 2. Root cause

`webtyp.com/view`'s `callerLister.list()` (the code every crudview-backed list runs on `Reload()`)
always calls the List op with **no args** — this is correct, by design: `router.Caller.Call`'s own
doc says args may be nil, and a List op is expected to tolerate an absent/empty args object (an
already-fixed `webtyp.com/mcp` bug, unrelated to this one, made sure "no args" now round-trips as
`{}` instead of corrupting the request — see that repo's own `docs/PLAN.md` history). So on every
real page load, `opListDevices` (`ops.go`) receives `ListDevicesArgs{TenantId: ""}`.

`ListDevices("", filter)` (`module.go`) then filters `WHERE tenant_id = ''` — which legitimately
matches nothing, silently, no error. The list renders "0 / 0" even for a tenant with real rows.

`github.com/veltylabs/staff_manager` already solves this exact problem, and this plan ports its
pattern exactly:

```go
// staff_manager/module.go
type Deps struct {
	IDs         model.IDGenerator
	Publisher   events.Publisher
	TenantID    string
	...
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("staff_manager: Deps.IDs is required")
	}
	if deps.TenantID == "" {
		return nil, fmt.Err("staff_manager: Deps.TenantID is required")
	}
	...
}

// staff_manager/ops.go
func (m *Module) opListStaff(ctx router.Context) {
	var args ListStaffArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	tenantID := args.TenantId
	if tenantID == "" {
		tenantID = m.tenantID
	}
	list, err := m.ListStaff(tenantID)
	...
}
```

`device_manager` has neither the field nor the fallback.

## 3. The fix

### `module.go`

1. Add `TenantID string` to `Deps`, next to `Publisher`:
   ```go
   type Deps struct {
   	IDs       model.IDGenerator // required — the module never builds its own
   	Publisher events.Publisher  // optional — nil disables publishing silently
   	TenantID  string            // required — this installation's tenant id, the
   	                            // fallback opListDevices uses when a caller sends
   	                            // no tenant_id (every crudview-backed list does)
   }
   ```
2. Add `tenantID string` to `Module`, next to `pub`.
3. In `New`, require it exactly like `IDs`:
   ```go
   func New(db *orm.DB, deps Deps) (*Module, error) {
   	if deps.IDs == nil {
   		return nil, fmt.Err("device_manager: Deps.IDs is required")
   	}
   	if deps.TenantID == "" {
   		return nil, fmt.Err("device_manager: Deps.TenantID is required")
   	}
   	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher, tenantID: deps.TenantID}, nil
   }
   ```

### `ops.go`

In `opListDevices`, between decoding args and calling `m.ListDevices`, add the fallback:

```go
func (m *Module) opListDevices(ctx router.Context) {
	var args ListDevicesArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	tenantID := args.TenantId
	if tenantID == "" {
		tenantID = m.tenantID
	}
	filter := DeviceFilter{Type: args.Type, ActiveOnly: args.ActiveOnly, Limit: args.Limit, Offset: args.Offset}
	devices, err := m.ListDevices(tenantID, filter)
	...
```

(the rest of the function body is unchanged — only the `tenantID` variable is introduced and used
in place of `args.TenantId` on the `m.ListDevices(...)` call).

### What NOT to do

- **Do not add this fallback to `opGetDevice`, `opDeactivateDevice`, or `opDeleteDevice`.**
  Those ops always receive an explicit `id` from a real UI action (a row click, a delete
  confirmation) — never through `callerLister.list()`'s args-less path — so they are not affected
  by this bug. `staff_manager` draws the identical line (only `opListStaff` falls back). Adding it
  everywhere "for consistency" would be scope creep past what this bug requires.
- **Do not touch `opCreateDevice`/`opUpdateDevice`/`opUpsertDevice`.** Those take a full `Device`
  record from the caller, including its own `tenant_id` — a create/update is never called with an
  empty tenant by design, and this bug is specific to the List op's args-less Reload path.
- **Do not touch `webtyp.com/view`.** `callerLister.list()` passing no args is correct per
  `router.Caller`'s own documented contract — the fix is this module adopting the fallback
  `staff_manager` already has, not changing the generic caller.
- **Do not make `Deps.TenantID` optional with a silent empty-string default.** Require it in `New`,
  exactly like `IDs` — a module silently running with an empty tenant fallback would scope every
  fallback query to `tenant_id = ''`, which is a worse, harder-to-diagnose version of today's bug.

## 4. Verification

```bash
gotest
# vet ✅, race ✅, tests ✅ — TestOpListDevices_FallsBackToModuleTenant now PASSES,
# and every other existing test (TestTenantIsolation, TestMountOperations_CreateDevice, etc.)
# still passes.
```

## 5. Downstream consumers (informational — not part of this plan's scope)

Once this ships as a new tagged version, `github.com/veltylabs/mjosefa-cms` needs: a version bump
(`go get github.com/veltylabs/device_manager@<new-version>`), AND its own
`modules/device_manager/server.go` wrapper updated to actually pass `tenantID` through to
`devicemanager.Deps{TenantID: tenantID}` (today that parameter is received and explicitly
documented as "unused"). Both are the consuming app's own job, tracked there, not here.

## Stages

| # | Stage | File(s) | Acceptance |
|---|---|---|---|
| 1 | Add `Deps.TenantID`, require it in `New` | `module.go` | Matches the exact shape shown above |
| 2 | Fall back to it in `opListDevices` only | `ops.go` | No other op touched |
| 3 | Verify | — | `gotest` green, `TestOpListDevices_FallsBackToModuleTenant` passes |
