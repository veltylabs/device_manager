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
		{devicemanager.ErrNotFound, "device not found"},
		{devicemanager.ErrIPAlreadyExists, "device ip already exists"},
		{devicemanager.ErrMACAlreadyExists, "network interface mac already exists"},
		{devicemanager.ErrRandomizedMAC, "randomized MAC address: on the device, turn off random (private) hardware addresses for this Wi-Fi network and register its real MAC"},
		{devicemanager.ErrNoZone, "ip is empty and the device has no zone to assign one from"},
		{devicemanager.ErrIPOutsideZone, "ip is outside the device's zone range"},
		{devicemanager.ErrZoneFull, "zone has no free ip"},
		{devicemanager.ErrZoneOverlap, "zone range overlaps another zone"},
		{devicemanager.ErrZoneInUse, "zone is used by a device"},
		{devicemanager.ErrZoneNotFound, "zone not found"},
		{devicemanager.ErrInterfaceNotFound, "network interface not found"},
		{devicemanager.ErrInvalidZoneRange, "zone range must be two IPv4 addresses with start <= end"},
	}

	for _, c := range cases {
		if got := c.err.Error(); got != c.expected {
			t.Errorf("Expected error string %q, but got %q", c.expected, got)
		}
	}
}
