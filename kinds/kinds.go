// Package kinds contiene los tipos cerrados para campos con opciones de device_manager.
// Viven en su propio paquete porque ormc resuelve el kind de un campo compilando
// y ejecutando su constructor, el cual solo puede importar desde un paquete independiente.
package kinds

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/network"
)

// Tipos de equipos — el ÚNICO lugar donde existen estos literales (reexportados por
// el paquete raíz como DeviceType*).
//
// Ejemplo: tipo := kinds.TypeComputer
const (
	TypeComputer = "computer"
	TypePrinter  = "printer"
	TypeServer   = "server"
	TypeOther    = "other"
)

// DeviceType es el control de radio con opciones cerradas para Device.type.
//
// Ejemplo: in := kinds.DeviceType()
func DeviceType() input.Input {
	return input.Radio(
		fmt.KeyValue{Key: TypeComputer, Value: "Computer"},
		fmt.KeyValue{Key: TypePrinter, Value: "Printer"},
		fmt.KeyValue{Key: TypeServer, Value: "Server"},
		fmt.KeyValue{Key: TypeOther, Value: "Other"},
	)
}

// DeviceAccess es el control de radio con opciones cerradas para Device.access; los valores son
// los nombres de acceso definidos en webtyp.com/network.
//
// Ejemplo: in := kinds.DeviceAccess()
func DeviceAccess() input.Input {
	return input.Radio(
		fmt.KeyValue{Key: network.AccessLocalName, Value: "Local network only"},
		fmt.KeyValue{Key: network.AccessInternetFilteredName, Value: "Internet (filtered)"},
		fmt.KeyValue{Key: network.AccessInternetName, Value: "Internet"},
	)
}
