---
PLAN: "fix: move schema creation out of New() into a separate migrate subpackage, matching webtyp.com/auth and webtyp.com/rbac"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — a `migrate/` subpackage replaces the `CreateTable` call inside `New()`

## Part of a multi-repo wave

This is module 2 of `DDL_MIGRATE_ISOLATION_MASTER_PLAN.md` (orchestrator:
`webtyp.com/app-releases`, `docs/DDL_MIGRATE_ISOLATION_MASTER_PLAN.md`). No
dependencies — dispatch any time, in parallel with `item_catalog`'s and
`clinical_encounter`'s identical plans. `veltylabs/mjosefa-cms`'s
`cmd/migrate` stage depends on the tag this plan produces.

## Why

Same defect as `item_catalog` (the reference module for this wave, see its
`docs/PLAN.md` at
`https://github.com/veltylabs/item_catalog/blob/main/docs/PLAN.md` for the
full justification): `New()` (`module.go:27-37`) runs
`ddl.CreateTable(&Device{})` unconditionally against a real SQL backend,
with no separate, callable migration step. This repo brings itself in line
with `webtyp.com/auth/authority.Migrate` / `webtyp.com/rbac.Migrate` — the
pattern `item_catalog` establishes for this whole wave.

**Migrate must live in its own subpackage, `migrate/`, not a new file in
the root `devicemanager` package.** `mjosefa-cms/modules/device_manager/
view.go` (compiled into the WASM client) imports this repo's root package
for `devicemanager.NewView(caller)`. If `Migrate`/`webtyp.com/ddl` stayed in
that same root package, `ddl` would still link into the WASM binary through
the view import, regardless of any build tag the composition root puts on
its own `backend.go`. A separate `migrate/` subpackage that nothing on the
WASM build path ever imports keeps `ddl` out of that build graph entirely.

**This plan does not merge or consolidate `device_manager` into
`item_catalog` or any other module — it stays a fully independent module,
with its own independent `Migrate`.**

## What to change

### 1. `module.go` — remove the DDL block from `New()`

Before:

```go
func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("device_manager: Deps.IDs is required")
	}
	if ddlCompiler, ok := db.RawConn().(ddl.Compiler); ok {
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&Device{}); err != nil {
			return nil, err
		}
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}
```

After:

```go
func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("device_manager: Deps.IDs is required")
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}
```

Also update the doc comment directly above `New` (currently explains the
type-assert-for-DDL idiom) — replace it with a plain one-liner, e.g. `// New
connects the module to an already-connected *orm.DB; the schema is assumed
to already exist — see the migrate subpackage.` Remove the now-unused
`"webtyp.com/ddl"` import from `module.go` (grep the file first: nothing
else in it uses `ddl.`).

### 2. New package `migrate/` (a subdirectory, not a file in the root package)

Create `migrate/migrate.go`:

```go
package migrate

import (
	"webtyp.com/ddl"

	devicemanager "github.com/veltylabs/device_manager"
)

// Migrate reconciles the database schema device_manager owns: Device.
//
// It is deliberately NOT called by New, and deliberately lives in its own
// package rather than a new file in the root package: nothing on a
// consuming app's WASM build path (its view.go, which imports the root
// devicemanager package for devicemanager.NewView) ever imports
// "github.com/veltylabs/device_manager/migrate" — so webtyp.com/ddl never
// enters that build graph, regardless of build tags on the consumer's side.
//
// conn is a ddl.Execer, not an *orm.DB, so a deploy-time transport that can
// only execute DDL satisfies it. An *orm.DB's RawConn() also satisfies it,
// for local/test callers:
//
//	conn, _ := postgres.Open(dsn)
//	compiler, _ := conn.(ddl.Compiler)
//	err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	return ddl.New(conn, ddlCompiler).CreateTable(&devicemanager.Device{})
}
```

Import path for consumers: `github.com/veltylabs/device_manager/migrate`.
The root package keeps its existing name (`devicemanager`, see
`module.go:1`); only the new subdirectory's package is called `migrate`.

### 3. New test file `migrate/migrate_test.go`

This test lives inside the new `migrate/` package directory, next to the
code — not under the root `tests/` directory, since `migrate` is its own
package (same convention `webtyp.com/auth/authority` uses for its own
`migrate_test.go`):

```go
package migrate_test

import (
	"testing"

	"github.com/veltylabs/device_manager/migrate"
	"webtyp.com/ddl"
	"webtyp.com/model"
)

type dummyExecer struct{ calls []string }

func (d *dummyExecer) Exec(query string, args ...any) error {
	d.calls = append(d.calls, query)
	return nil
}

type dummyCompiler struct{}

func (d *dummyCompiler) CompileDDL(stmt ddl.Stmt, m model.Model) (string, []any, error) {
	return stmt.Table, nil, nil
}

func TestMigrate_CreatesDeviceTable(t *testing.T) {
	execer := &dummyExecer{}
	if err := migrate.Migrate(execer, &dummyCompiler{}); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	if len(execer.calls) != 1 {
		t.Fatalf("got %d Exec calls, want 1: %v", len(execer.calls), execer.calls)
	}
}
```

### 4. This repo's own `AGENTS.md` — replace the "Persistence" bullet

Replace the current bullet (search for `**Persistence**:`) with this exact
block (from `DDL_MIGRATE_ISOLATION_MASTER_PLAN.md` §2,
`https://github.com/webtyp/app-releases/blob/main/docs/DDL_MIGRATE_ISOLATION_MASTER_PLAN.md`):

```markdown
- **Persistence**: `New(db *orm.DB, deps Deps)` receives an already-connected
  `*orm.DB` and assumes its schema already exists — it never creates or
  alters tables, and never imports `webtyp.com/ddl`. Schema reconciliation
  lives in its own **subpackage**, `<module>/migrate` (`package migrate`),
  exporting `Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error`.
  Deliberately not called by `New`, and deliberately not in the module's
  root package: schema work is deploy-time work, run once from a migration
  binary (`cmd/migrate` in the composition-root app) — and keeping it in a
  separate package means nothing on a WASM build's import path (`view.go`,
  `init.go`, the root package itself) ever pulls `webtyp.com/ddl` into that
  binary, regardless of build tags.
  ```go
  // migrate/migrate.go
  package migrate

  import (
      "webtyp.com/ddl"

      thismodule "github.com/veltylabs/<this-module>"
  )

  func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
      d := ddl.New(conn, ddlCompiler)
      if err := d.CreateTable(&thismodule.CatalogItem{}); err != nil {
          return err
      }
      return nil
  }
  ```
  A module's own tests build `*orm.DB` over `storage/mem`
  (`orm.New(mem.New())`), which creates tables lazily on first `Exec` — they
  never call `Migrate`, and `New` never needs to type-assert for
  `ddl.Compiler` at all anymore.
```

## What this does NOT change

- No change to `Deps`, `Module`, `DeviceFilter`, or any exported method
  other than the new `Migrate` — every existing caller of `New` keeps
  compiling unchanged.
- `storage/mem`-backed tests are unaffected, same reasoning as `item_catalog`.
- This module stays fully independent — no shared table, no shared
  `Migrate` call, no import relationship with `item_catalog` or
  `clinical_encounter`.

## Acceptance

- `grep -n "ddl\." module.go` → empty.
- `grep -rn "webtyp.com/ddl" *.go` (root package only, not `migrate/`) →
  empty.
- `gotest` passes, including `TestMigrate_CreatesDeviceTable`.
- `grep -n "Persistence" AGENTS.md` shows the new block.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `module.go` | DDL block removed from `New`, unused `ddl` import dropped |
| 2 | `migrate/migrate.go` (new package) | `Migrate` implemented exactly as specified, in its own subdirectory |
| 3 | `migrate/migrate_test.go` (new) | test passes |
| 4 | `AGENTS.md` | "Persistence" bullet replaced verbatim |
