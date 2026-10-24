---
PLAN: "feat(ui): one icon per screen from webtyp/icons, screens explain themselves, demo at web/browser/main.go"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# PLAN — the device screens: right icons and their own help

## Why

1. `ui/svg.go` draws an **apple** under a comment that says "laptop", and the three screens
   (Equipos, Zonas, Interfaces de red) all reuse that one glyph. The shared set `webtyp/icons`
   v0.0.33 now has the right glyphs: `laptop`, `location`, `network`.
2. No screen explains what it is for. The shell (`webtyp.com/layout/platformd`) draws a «?» button
   and a description for any screen wrapped with `platformd.Describe` (rule of the ecosystem: a
   module explains itself). `veltylabs/service_catalog` and `pharmacy_manager` already do it.
3. The demo lives at `web/browser.go`; the ecosystem layout is one `main` per folder:
   `web/browser/main.go` (no destination tag). The demo is how a reviewer looks at the screens.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/icons@v0.0.33
```

Do not change any other dependency version.

## APIs used

| Symbol | Exact signature | Tag / file |
|---|---|---|
| `laptop.Ref`, `location.Ref`, `network.Ref` | `const Ref = svg.Icon("laptop")` / `"location"` / `"network"` (type `webtyp.com/svg.Icon`) | `webtyp.com/icons v0.0.33`, `<glyph>/<glyph>.go` |
| `laptop.Def()`, `location.Def()`, `network.Def()` | `func Def() sprite.Definition` — behind `//go:build !wasm` | same, `<glyph>/svg.go` |
| `sprite.NewSprite` | `func NewSprite(defs ...sprite.Definition) *sprite.Sprite` | `webtyp.com/svg/sprite` (current `go.mod`) |
| `platformd.Describe` | `func Describe(m UIModule, text lang.Text, steps ...Step) UIModule` — keeps every capability of `m` (its `Activate` included) | `webtyp.com/layout v0.3.64`, `platformd/capabilities.go` |
| `platformd.Described` | `interface { Description() lang.Text; Steps() []Step }` | same |
| `platformd.NewUIModule` | `func NewUIModule(id, label string, icon svg.Icon, view Component, onActivate ...func()) UIModule` | `platformd/factory.go` |
| `lang.Text` | `type Text string` | `webtyp.com/lang v0.1.9` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR;
never adapt it by guessing.

## Stage 1 — `ui/svg.go` (rewrite the whole file)

```go
//go:build !wasm

package ui

import (
	"webtyp.com/icons/laptop"
	"webtyp.com/icons/location"
	"webtyp.com/icons/network"
	"webtyp.com/svg/sprite"
)

// Icons entrega al build los dibujos de las tres pantallas, tomados del
// conjunto compartido webtyp/icons: Equipos (laptop), Zonas (location) e
// Interfaces de red (network).
//
// Ejemplo: sprite := (&ui.Icons{}).IconSvg()
type Icons struct{}

// IconSvg devuelve los tres dibujos en un solo sprite.
//
// Ejemplo: sprite := (&ui.Icons{}).IconSvg()
func (m *Icons) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(laptop.Def(), location.Def(), network.Def())
}
```

This deletes `IconID()` (nothing calls it: `grep -rn "IconID()" .` → empty after the change) and
the apple path (`grep -rn "M350.85" .` → empty).

## Stage 2 — `ui/browser.go`

1. Imports: remove `"webtyp.com/svg"`; add `"webtyp.com/icons/laptop"`, `"webtyp.com/icons/location"`,
   `"webtyp.com/icons/network"` and `"webtyp.com/lang"`.
2. Below the `fieldZoneID`/`fieldDeviceID` block add exactly:

```go
// Textos de ayuda que el chasis muestra en el botón «?» de cada pantalla. Las
// claves están en inglés; la traducción vive en lang.json.
const (
	descDevices    lang.Text = "Each computer, printer or device of the clinic, with the zone where it is. Its network interfaces are added in Network interfaces."
	descZones      lang.Text = "Zones group the devices by place (for example, reception or box 1). Create the zones first; each device then picks one."
	descInterfaces lang.Text = "The network connections of each device: its MAC and IP address. Staff log in from the devices registered here."
)
```

3. The three `return` lines become:

```go
	return platformd.Describe(platformd.NewUIModule(ID, Label, laptop.Ref, v, activate), descDevices), nil
```
```go
	return platformd.Describe(platformd.NewUIModule(ZonesID, ZonesLabel, location.Ref, v), descZones), nil
```
```go
	return platformd.Describe(platformd.NewUIModule(InterfacesID, InterfacesLabel, network.Ref, v, activate), descInterfaces), nil
```

Nothing else in the file changes (keep its `//go:build browser` line).

## Stage 3 — `lang.json` (new)

Run `langc sync .`. It creates `lang.json`. Fill every empty **Spanish** (`es`) value; for the three
descriptions use exactly:

```json
    "Each computer, printer or device of the clinic, with the zone where it is. Its network interfaces are added in Network interfaces.": ["Cada computador, impresora o equipo de la clínica, con la zona donde está. Sus interfaces de red se agregan en Interfaces de red."],
    "The network connections of each device: its MAC and IP address. Staff log in from the devices registered here.": ["Las conexiones de red de cada equipo: su dirección MAC e IP. El personal inicia sesión desde los equipos registrados aquí."],
    "Zones group the devices by place (for example, reception or box 1). Create the zones first; each device then picks one.": ["Las zonas agrupan los equipos por lugar (por ejemplo, recepción o box 1). Crea primero las zonas; luego cada equipo elige una."],
```

Other languages `langc` adds may stay empty. List in the PR every other key you translated.

## Stage 4 — demo: `web/browser.go` → `web/browser/main.go`

Create `web/browser/main.go` with the **same content** as `web/browser.go`, changing only the first
line: `//go:build browser` → `//go:build wasm`. Then delete `web/browser.go`.
Check: `test ! -e web/browser.go && test -e web/browser/main.go`.

## Stage 5 — test `tests/ui_described_test.go` (new)

```go
//go:build browser

package tests

import (
	"testing"

	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/layout/platformd"
)

// Las tres pantallas se explican solas y cada una tiene su propio dibujo.
func TestUI_ScreensAreDescribedWithTheirOwnIcon(t *testing.T) {
	fake := &uiFakeCaller{}
	devices, err := ui.Browser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatal(err)
	}
	zones, err := ui.ZonesBrowser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatal(err)
	}
	ifaces, err := ui.InterfacesBrowser(fake, &mockIDGen{}, "tenant-test")
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		screen platformd.UIModule
		icon   string
	}{{devices, "laptop"}, {zones, "location"}, {ifaces, "network"}}
	for _, w := range want {
		d, ok := w.screen.(platformd.Described)
		if !ok || d.Description() == "" {
			t.Errorf("%s must implement platformd.Described with a non-empty text", w.screen.ModelName())
		}
		if got := w.screen.Icon().ID(); got != w.icon {
			t.Errorf("%s icon = %q, want %q", w.screen.ModelName(), got, w.icon)
		}
	}
}
```

`uiFakeCaller` (`tests/ui_view_test.go`) and `mockIDGen` (`tests/setup_test.go`) already exist;
`UIModule.Icon() svg.Icon` is declared in `webtyp.com/layout v0.3.64` `platformd/platformd.go:99`.
The existing tests in `tests/ui_view_test.go` must keep passing unchanged.

## Stage 6 — language

Translate to Spanish every README, `docs/*` file, diagram and code comment of this repo that is
still in English (only those; never rewrite what is already Spanish). Add a short `// Ejemplo: …`
line to every exported symbol's comment that lacks one. In `README.md`, the part about `ui/` says
that each screen has its own icon from `webtyp/icons` and a description (`platformd.Describe`), and
that the demo is `web/browser/main.go`.

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission or report progress: when the check passes, commit, push and open
  the PR. Stuck for 30 minutes: open the PR anyway listing what fails.
- Never edit a generated file (`model_orm.go`).
- Commit after every stage (`git add -A` then `git commit -m "stage N: …"`); never `git reset`,
  `git restore`, `git checkout -- <file>` or `git stash` on your own work.
- `git diff --stat main` must list every file of the stages and nothing the plan did not order.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `ui/svg.go` | apple path and `IconID` gone; three shared glyphs |
| 2 | `ui/browser.go` | three screens described, each with its own `Ref` |
| 3 | `lang.json` (new) | no empty `es` value |
| 4 | `web/browser/main.go` (new), `web/browser.go` (deleted) | demo at the new path |
| 5 | `tests/ui_described_test.go` (new) | test passes |
| 6 | `README.md`, `docs/*`, comments | no English comment or doc left |
