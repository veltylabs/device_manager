//go:build browser

package module

import (
	"webtyp.com/appkit"

	"github.com/veltylabs/device_manager/ui"
)

// En el navegador el módulo solo aporta sus pantallas.
var server appkit.ModuleServer

var browser = appkit.ModuleBrowser{Screens: ui.Screens}
