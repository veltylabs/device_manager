package devicemanager

import "webtyp.com/fmt"

const ipv4Octets = 4

// ipv4ToUint32 parses a dotted IPv4 address into a number, so ranges compare
// numerically. ok=false for anything that is not a dotted IPv4.
func ipv4ToUint32(ip string) (uint32, bool) {
	parts := fmt.Convert(ip).Split(".")
	if len(parts) != ipv4Octets {
		return 0, false
	}
	var n uint32
	for _, p := range parts {
		if p == "" {
			return 0, false
		}
		v, err := fmt.Convert(p).Int()
		if err != nil || v < 0 || v > 255 {
			return 0, false
		}
		n = n<<8 | uint32(v)
	}
	return n, true
}

// uint32ToIPv4 is the inverse of ipv4ToUint32.
func uint32ToIPv4(n uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", n>>24, (n>>16)&0xff, (n>>8)&0xff, n&0xff)
}

// zoneRange returns a zone's range as numbers; ok=false when it is not a
// valid IPv4 range with start <= end.
func zoneRange(z Zone) (start, end uint32, ok bool) {
	s, ok1 := ipv4ToUint32(z.RangeStart)
	e, ok2 := ipv4ToUint32(z.RangeEnd)
	if !ok1 || !ok2 || s > e {
		return 0, 0, false
	}
	return s, e, true
}
