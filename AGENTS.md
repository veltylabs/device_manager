# Workflow configuration

<!-- webtyp:base — copied from devskills skill agents-md; do not edit here -->
## Ecosystem rules — read before any change

`docs/PLAN.md` is the spec: no design judgment; anything unclear is a question in the PR, not a guess.

- Public API: typed (no `any`/`interface{}`), explicit, minimal; never export a symbol only a test uses.
- One intent, one path: the old path is deleted in the same change. No silent fallback, `TODO`, commented code or stub.
- Fail at compile time, else a loud diagnostic — never silently. Zero values deny.
- A defect in a dependency is reported, never wrapped or patched here.
- No string literals in logic: named constants. The same logic never lives in two repos: import the package the plan names; if you find an existing one the plan does not name that already does what you are about to write, say so in the PR instead of writing it.
- `cmd/*/main.go`: parse flags, inject, print/exit only. stdout = data, stderr = logs, exit codes are the contract.
- WASM code: `webtyp/fmt`, `webtyp/time`, `webtyp/json` instead of stdlib; no `map`; `dom.Element` embedded by value.
- `css.go`, `svg.go`, `js.go`, `html.go` carry `//go:build !wasm`; no `ssr.go`, no `front.go`.
- Build tags: `wasm`/`!wasm` only when the API does not exist on that platform; `browser`/`!browser` when an app's file belongs to another binary; a library carries no destination tag (the linker drops what a binary never calls). One main per folder, no destination tag: `web/browser/main.go` → `web/public/browser.wasm`, `web/server/main.go` → `web/server/server`, `web/edge/main.go` → `web/edge/edge.wasm`.
- Code and keys in English; UI text is `lang.Text`, translations live in `lang.json`. Comments and docs: English in `webtyp/*`; in `veltylabs/*` **every README, `docs/*` file, diagram and code comment is in Spanish**, written for a reader with little technical background, and each exported symbol's comment shows a short usage example.
- A translation key is reused, never duplicated as a near-identical phrase (`"not a valid integer"` + `"not a valid decimal number"` repeats `"not a valid"` three ways instead of composing it once) — every library ships to WASM/TinyGo, so each unreused phrase is dictionary rows AND binary bytes, in every consumer, forever. A token that doesn't need translating (a kind/type name like `int`) is passed as data (a variable), never as a string literal — a literal becomes a dictionary key whether or not it needed one.
- Tests cover behaviour and domain rules, never the look (CSS, colours, markup order, class strings). In `tests/` (external package), real stack with fakes only at the edge; run `gotest`.
- Docs (`README.md`, `docs/*`) updated with the API; permanent docs never link `docs/PLAN.md`.
- Before the PR: `git diff --stat main` lists every file of the plan's stages, and nothing the plan did not order is deleted.
<!-- /webtyp:base -->

## This repo

-   **Workflow status**: **Active**
-   **Architecture**: Domain module (`veltylabs/*`) — implements `router.OperationModule`.
-   **Composition root**: Never runs alone; instantiated and wired by `veltylabs/app`.
-   **Current focus**: Data modeling, rules, and tests.

---

# Architecture bounds and code constraints

This is a **domain module** repository. The root package (the module itself) adheres to strict constraints to remain technology-agnostic and universally composable:

## Language and ecosystem

-   **Go only**: This repository contains pure Go code. No shell scripts, Makefiles, JS/TS, Python, Dockerfiles, or CI/CD configuration files are permitted.
-   **Language**: `es` (Spanish) for docs, comments, UI text and struct field descriptions. `en` (English) for exported Go identifiers, JSON keys, SQL tables/columns, log messages, and test names.

## The whitelist — what the root package may import

The root package (`*.go` in the repository root) may ONLY import from the standard library (subject to the blacklist below) and:
- `webtyp.com/fmt`, `webtyp.com/input`, `webtyp.com/events`, `webtyp.com/network`, `webtyp.com/time` (domain primitives and utilities).
- `webtyp.com/model` (entities, validation, definitions, ops).
- `webtyp.com/orm` (persistence via generated typed query builders).
- `webtyp.com/router` (transport/RPC, ops registration, typed RPC client).
- `webtyp.com/view` (presentation logic decoupled from DOM/React/Vue).
- `webtyp.com/lang` (translations).

## The blacklist — what the root package MUST NEVER import

The root package must remain entirely independent of rendering engines, UI frameworks, protocol handlers, direct DOM manipulation, and explicit storage engines.
- **Never import `webtyp.com/layout/*` or `webtyp.com/components/*`**. UI logic (React/Vue/DOM equivalents) belongs strictly in the `ui/` subpackage.
- **Never import `webtyp.com/storage/*`**. The module accepts an already-connected `*orm.DB`.
- **Never import `webtyp.com/dom`, `webtyp.com/css`, `webtyp.com/widget`, `webtyp.com/html`, `webtyp.com/svg`**.
- **Never import standard library formatting/conversion/reflection**: `fmt`, `strconv`, `reflect` are strictly banned. Use `webtyp.com/fmt` instead.
- **Never import standard library time**: `time` is strictly banned. Use `webtyp.com/time` instead.
- **Never import other domain modules**: A root module must never import another domain module directly. Cross-module wiring is done via interfaces at the composition root.
- **No `replace` directives**: A `go.mod` file must never contain `replace` directives (especially local paths). Use `go get module@version`.

## Go idioms and rules

- **Interface equality**: Direct interface equality comparisons (e.g., `err == ErrNotFound`), `switch` on interface values, and `errors.Is`/`errors.As` are strictly prohibited to maintain TinyGo/WASM compatibility. Always use type assertions on unexported custom types (e.g., `if e, ok := err.(domainError); ok && e == ErrNotFound`).
- **Sentinel Errors**: For external errors, use provided query functions (e.g., `orm.IsNotFound(err)`). For internal sentinel errors, define them as constants of an unexported string alias:
  ```go
  type domainError string
  func (e domainError) Error() string { return string(e) }
  const ErrNotFound domainError = "not found"
  ```
- **Maps are prohibited**: The use of Go maps (`map[K]V`) is strictly prohibited. Use slices or structs.

## Data definitions (`model.Definition`) and generated code

- Types implementing `model.Definition` (the domain entities) are annotated with field tags (`label`, `desc`, `form`, `type`, `limit`, etc.).
- When definitions change, the schema and queries must be regenerated using `ormc` (`$(go env GOPATH)/bin/ormc`).
- `ormc` strictly generates `model_orm.go`. It does not accept parameters.
- **CRITICAL**: `ormc` alters `go.mod`. You must backup `go.mod` before running `ormc`, run it, and restore `go.mod`, followed by `go mod tidy`.
  ```bash
  cp go.mod /tmp/go.mod.bak && ormc && cp /tmp/go.mod.bak go.mod && go mod tidy
  ```
- Types implementing `model.Definition` must be defined in the root package, never in subpackages. `ormc` does not scan subpackages. Custom closed-set types (e.g., enums/radios) must live in a subpackage (e.g., `kinds/`) to prevent `ormc` from trying to generate database tables for them.

## Operations (`router.OperationModule`)

- The module exposes its capabilities via `MountOperations(reg router.OperationRegistry)`.
- It registers typed operations using `reg.Register(OpName, handler)`.
- **`handler.Requires(resource, action)`**: Every operation MUST declare the exact permissions it requires. If an operation performs both a read and a write, it requires both (e.g., `model.ActionRead|model.ActionUpdate`).
- **`handler.Accepts(args)`**: Every operation MUST declare the payload it accepts. If it accepts no arguments, it MUST explicitly declare `.Accepts(nil)`. Do not invent empty `struct{}` argument types.
- Never swallow actual database errors as `not found`. Only `orm.ErrNotFound` should be mapped to the domain's "not found" sentinel; anything else propagates as the internal error it is.

## Multi-tenancy

- Every tenant-scoped table MUST carry `tenant_id` (`NotNull: true`).
- **Every UPDATE/DELETE condition MUST include the tenant column.** `orm.Eq(X_.Id, id)` alone is a cross-tenant write vulnerability. The pattern of fetching with a tenant check, and then mutating by ID only, introduces a TOCTOU (Time-of-Check Time-of-Use) window.
- If a tenant-scoped fetch fails, return the error. Do not implement "stub delete" fallbacks.
- Modules that are deliberately not tenant-scoped must state this explicitly in their "Domain-specific notes" and `docs/ARCHITECTURE.md`.

## Testing

- **Runner**: All tests MUST be run using `gotest` (`$(go env GOPATH)/bin/gotest`), never `go test` directly. (If missing, install via `go install github.com/webtyp/devflow/cmd/gotest@latest`).
- **WASM Compatibility**: Modules must compile to WASM. Always verify with: `GOOS=js GOARCH=wasm go vet ./...`
- **Scope**: Tests must live in `tests/` (package `tests`) and only exercise the exported API. `tests/` is a plain directory inside the root module, NEVER a nested Go module (no `tests/go.mod`).
- **Mocking**: Tests build their `*orm.DB` over `storage/mem` (`orm.New(mem.New())`). Drive `MountOperations` against `router/mock`. Test `view.Presenter` against `view/conformance`'s `FakeCaller` or a hand-rolled caller. Never use a concrete DB, transport, or renderer.
- **Tenant Isolation Tests**: Every tenant-scoped module MUST include targeted tests verifying that Tenant A cannot read/update/delete Tenant B's rows through any service method or operation.
- **Meaningful Assertions**: Tests must assert on results and behaviors. Calling methods and discarding returns (`_ = m.Validate(0)`) inflates coverage without proving correctness and is invalid.
- **Widget Regression**: Modules with form widgets in their Definitions must include a test verifying that `form.New(id, &GeneratedArgs{})` yields exactly the expected inputs.

## ui/, seed/ and web/ — the module's own view and demo

The whitelist and blacklist above apply to the **root domain package**. Three sub-packages are exempt, and only them:

- `ui/` (package `ui`, `ui/browser.go` tagged `browser`; `css.go`/`svg.go` tagged `!wasm`) may import `webtyp.com/icons/*`. It is the module's screen definition: `const ID`, `const Label` and `Screens(caller router.Caller) []view.Screen`.
- `seed/` (package `seed`, no build tag) holds demo data: `Load(...) (Data, error)`, which writes through the module's own methods so every row is validated, and returns the rows it created so downstream demos can reference them.
- `web/` (package `main`, `web/browser.go` tagged `browser`) is the runnable demo and may additionally import the concrete `webtyp.com/storage/mem`, `webtyp.com/router/loopback`, `webtyp.com/events/mock`, `webtyp.com/unixid` and `webtyp.com/auth/trusted_ip` (for `ValidateRUT`).

**Dependencies between modules form a DAG.** The root domain package still imports no sibling module. `ui/`, `seed/` and `web/` may import the domain, `ui/` and `seed/` packages of **upstream** modules only. Current graph: `device_manager`, `item_catalog`, `patient_directory`, `business_calendar`, `payment_manager` (leaves) ← `staff_manager` ← `appointment_booking` ← `clinical_encounter` (`appointment_booking` also depends on `item_catalog`, `patient_directory`, `business_calendar`, `payment_manager`; `clinical_encounter` also on `patient_directory` and `staff_manager`). A screen that mixes modules lives in the most-downstream module it touches.

## Publishing / dispatch

Changes to a `veltylabs/modules/*` repo go through the CodeJob workflow (see skill
**agents-workflow**): write a self-contained `docs/PLAN.md` with frontmatter (`PLAN`/`TAG`/
`EXECUTOR`/`REVIEWER`), the human dispatches it (`codejob`), an executor agent opens a PR, an
optional reviewer agent judges it, and closing the loop (`codejob 'msg'` or merging the PR) calls
`gopush` internally . A planning/authoring agent **writes** the plan; it
never runs `codejob` or `gopush` itself — dispatch and close are the human's call.

## Documentation a module must carry

| File | Purpose |
|---|---|
| `AGENTS.md` | This file, copied verbatim + a "Domain-specific notes" section below the line |
| `docs/ARCHITECTURE.md` | Domain scope, entities, the patterns above applied to this module, Ops table, composition-root example |
| `docs/PLAN.md` | Present only while a change is in flight — self-contained, deleted once merged (never left around "for reference") |
| `docs/diagrams/database.md` | Mermaid ERD |
| `README.md` | Quick start, Ops table, key files |

---

## Domain-specific notes (edit per module — nothing above this line)

- **Domain**: inventory of a site's equipment — devices, their network interfaces (MAC + IP), zones
  (one per access point, with an IP range) and each device's network access level. Source of truth
  for `veltylabs/network_manager`; never talks to a router. Design: `docs/ARCHITECTURE.md`.
- **Router boundary**: `NetworkHosts` implements `webtyp.com/network`'s `HostSource` and
  `HostImporter`. Only interfaces **with a MAC** of active devices become hosts.
- **Login by IP**: `FindByIP` / `IPLocator` resolve an IP through `network_interface`; their
  signatures are relied on by `staff_manager`, `appointment_booking`, `clinical_encounter` — keep them.
  `seed.Load(m, tenantID)` is also called by those modules — keep its signature.
- **MACs**: canonical (`input.CanonicalMAC`), unique per tenant, randomized ones rejected.
- **Custom kinds** (`type`, `access` radios) live in `kinds/`: ormc rejects a kind constructor declared
  in the package it generates for. Never declare one in the root package again.
- **Tenant-scoped**: every row carries `tenant_id`; IP and MAC uniqueness are **per tenant**.
- **Publishes events**: `device_manager.device.*`, `device_manager.zone.*`,
  `device_manager.network_interface.*` — `Deps.Publisher` optional, `nil` disables silently.
