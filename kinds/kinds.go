// Package kinds holds device_manager's closed-option field kinds. They live in
// their own package because ormc resolves a field's kind by compiling and
// running its constructor, which it can only import from a package of its own.
package kinds

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/network"
)

// Device types — the ONLY place these literals exist (re-exported by the root
// package as DeviceType*).
const (
	TypeComputer = "computer"
	TypePrinter  = "printer"
	TypeServer   = "server"
	TypeOther    = "other"
)

// DeviceType is the closed-options radio for Device.type.
func DeviceType() input.Input {
	return input.Radio(
		fmt.KeyValue{Key: TypeComputer, Value: "Computer"},
		fmt.KeyValue{Key: TypePrinter, Value: "Printer"},
		fmt.KeyValue{Key: TypeServer, Value: "Server"},
		fmt.KeyValue{Key: TypeOther, Value: "Other"},
	)
}

// DeviceAccess is the closed-options radio for Device.access; the values are
// webtyp.com/network's access names.
func DeviceAccess() input.Input {
	return input.Radio(
		fmt.KeyValue{Key: network.AccessLocalName, Value: "Local network only"},
		fmt.KeyValue{Key: network.AccessInternetFilteredName, Value: "Internet (filtered)"},
		fmt.KeyValue{Key: network.AccessInternetName, Value: "Internet"},
	)
}
