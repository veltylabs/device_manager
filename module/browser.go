//go:build browser

package module

import (
	"webtyp.com/view"

	"github.com/veltylabs/device_manager/ui"
)

// En el navegador el módulo solo aporta sus pantallas.

var browser = view.ModuleBrowser{Screens: func(d view.BrowserDeps) []view.Screen { return ui.Screens(d.Caller) }}
