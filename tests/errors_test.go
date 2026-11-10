package tests

import (
	"testing"

	"github.com/veltylabs/device_manager"
)

func TestSentinelErrorsStrings(t *testing.T) {
	cases := []struct {
		err      error
		expected string
	}{
		{devicemanager.ErrNotFound, "device no encontrado"},
		{devicemanager.ErrIPAlreadyExists, "ip duplicada"},
		{devicemanager.ErrMACAlreadyExists, "mac duplicada"},
		{devicemanager.ErrRandomizedMAC, "dirección MAC aleatoria: apaga las direcciones privadas"},
		{devicemanager.ErrNoZone, "ip vacía y sin zona"},
		{devicemanager.ErrIPOutsideZone, "ip fuera de rango de zona"},
		{devicemanager.ErrZoneFull, "zona sin ip libre"},
		{devicemanager.ErrZoneOverlap, "rango solapado"},
		{devicemanager.ErrZoneInUse, "zona en uso"},
		{devicemanager.ErrZoneNotFound, "zona no encontrada"},
		{devicemanager.ErrInterfaceNotFound, "interfaz de red no encontrada"},
		{devicemanager.ErrInvalidZoneRange, "rango inválido"},
	}

	for _, c := range cases {
		if got := c.err.Error(); got != c.expected {
			t.Errorf("Expected error string %q, but got %q", c.expected, got)
		}
	}
}
