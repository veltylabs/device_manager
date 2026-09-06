---
PLAN: "refactor!: migrate github.com/tinywasm -> webtyp.com + adopt view.NewCallerLister"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 12809586858990536772
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `device_manager`: WebTyp rename + new `view.New` API

The framework moved from `github.com/tinywasm/*` to the vanity path
`webtyp.com/*`, and **every framework module is now published** under the new
path. `origin/main` of this repo is still entirely on `github.com/tinywasm/*`.
Two jobs:

- **A. The mechanical rename** `github.com/tinywasm/*` → `webtyp.com/*`.
- **B.** Adopt the new `webtyp.com/view` `view.New` signature.

The module import path **stays** `github.com/veltylabs/device_manager`.

---

## A. Rename `github.com/tinywasm` → `webtyp.com`

### A1. Go source (`*.go`, including any `tests/*.go`)

Replace import-path prefix **`github.com/tinywasm/`** → **`webtyp.com/`** in
every `.go` file. Grep to find them all:

```
grep -rln 'github.com/tinywasm' --include='*.go' .
```

Package selectors do not change — only the path in the `import` block.

### A2. `go.mod`

`origin/main` `require` block:

```
github.com/tinywasm/ddl v0.0.4
github.com/tinywasm/events v0.0.2
github.com/tinywasm/fmt v0.25.5
github.com/tinywasm/form v0.3.13
github.com/tinywasm/model v0.1.2
github.com/tinywasm/orm v0.11.1
github.com/tinywasm/router v0.1.15
github.com/tinywasm/storage v0.0.2
github.com/tinywasm/time v0.5.0
github.com/tinywasm/view v0.1.1
github.com/tinywasm/json v0.5.17 // indirect
```

For **each** `github.com/tinywasm/<X>`:

```
go mod edit -droprequire=github.com/tinywasm/<X>
go get webtyp.com/<X>@latest
```

Then `go mod tidy`. `@latest` is authoritative; current published tags for
reference: `ddl v0.0.15`, `events v0.0.3`, `fmt v1.0.0`, `form v0.4.7`,
`model v0.1.8`, `orm v0.12.1`, `router v0.1.31`, `storage v0.0.7`,
`time v0.5.5`, `view v0.5.2`, `json v0.5.25`. No `github.com/tinywasm/*` left in
`go.mod`; no `replace … => ../…` pointing outside this module.

### A3. Docs / config text

`*.md` / `*.yml` / `*.yaml`: `github.com/tinywasm/` → `github.com/webtyp/`.
Prose `TinyWasm`/`TinyWASM` → `WebTyp`. Leave `LICENSE` and upstream "TinyGo"
references untouched.

---

## B. New `view.New` API — use `view.NewCallerLister`

### The change

`webtyp.com/view`'s `view.New` no longer takes a `router.Caller`, an op-name
string, a slice factory, or `view.WithSaveOp` / `view.WithDeleteOp` (both
**removed**):

```go
func New(l Lister, record model.Model, opts ...Option) Presenter
type Lister interface{ List() ([]model.Model, error) }
```

The framework ships the adapter that reproduces the old op/caller behaviour
exactly:

```go
func NewCallerLister(c router.Caller, ops Ops, newList func() model.ModelSlice) Lister
type Ops struct{ List, Save, Update, Delete string }  // Ops.List required;
// the returned Lister carries exactly the write caps whose op name is non-empty
```

### Reference implementations (read first)

- **`webtyp.com/auth` → `auth/view.go`** (canonical):
  ```go
  b := view.NewCallerLister(caller,
      view.Ops{List: OpListUsers, Save: OpUpsertUser, Delete: OpDeleteUser},
      func() model.ModelSlice { return &UserList{} })
  return view.New(b, &User{}, view.WithTitle("Usuarios"))
  ```
- **`github.com/veltylabs/business_hours`** `main` — the same migration, just
  completed on a sibling.

### Do — rewrite `view.go`

`NewView` keeps its **exact current signature**
(`func(caller router.Caller) view.Presenter`). All three imports
(`webtyp.com/model`, `webtyp.com/router`, `webtyp.com/view`) stay. Body:

```go
// NewView builds the device Presenter — the tech-agnostic engine a renderer
// (crudview, or any other) wraps. This module builds it (view + model + router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListDevices, Save: OpUpsertDevice, Delete: OpDeleteDevice},
		func() model.ModelSlice { return &DeviceList{} })
	return view.New(b, &Device{}, view.WithTitle(titleDevices))
}
```

`titleDevices` is a new unexported constant in this package
(`const titleDevices = "Equipos"`) — do not inline the literal. `OpListDevices`,
`OpUpsertDevice`, `OpDeleteDevice` already exist in `ops.go` — reuse them.
`(*Device).Item()` stays exactly as it is.

### Tests

`NewView`'s signature is unchanged, so tests calling `NewView(fakeCaller)` keep
working. Adapt any test that references a removed symbol
(`view.WithSaveOp` / `view.WithDeleteOp` / old `view.New` arity) minimally,
preserving the assertion intent (lists rows, is Saver, is Deleter).

---

## Verify

```
grep -rn 'github.com/tinywasm' --include='*.go' --include='go.mod' .   # empty
grep -rn 'view.WithSaveOp\|view.WithDeleteOp' .                        # empty
grep -rn '=> \.\./' go.mod                                            # empty
```

## Acceptance

- `go build ./...` → clean.
- `gotest ./...` → all green (vet, race, tests).
- No `github.com/tinywasm/*` anywhere in `*.go` / `go.mod`.
- `view.go` no longer references `view.WithSaveOp` / `view.WithDeleteOp` / the
  5-arg `view.New`.
- `NewView` still has signature `func(caller router.Caller) view.Presenter`.

## Constraints

- **No behaviour change** — path rename + adapter swap only. The `view.Ops` op
  names are the same strings the old options used.
- **No hardcoded strings** — the view title is a named constant; op names
  already are (`ops.go`).
- Keep every `//go:build` tag exactly as-is (backend + WASM shared).
- Do not populate `view.Ops.Update` — the old view had no update op.
