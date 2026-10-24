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
