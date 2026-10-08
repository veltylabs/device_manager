package devicemanager

import "testing"

func TestStatusFor(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{ErrNotFound, 404},
		{ErrZoneNotFound, 404},
		{ErrInterfaceNotFound, 404},
		{ErrIPAlreadyExists, 409},
		{ErrMACAlreadyExists, 409},
		{ErrZoneOverlap, 409},
		{ErrZoneInUse, 409},
		{ErrZoneFull, 409},
		{nil, 500},
	}

	for _, c := range cases {
		if got := statusFor(c.err); got != c.expected {
			t.Errorf("statusFor(%v) = %d, expected %d", c.err, got, c.expected)
		}
	}
}
