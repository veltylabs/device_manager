package devicemanager

import "webtyp.com/crud"

// StoreDevice is the store of device.
func (m *Module) StoreDevice() crud.Store { return m.store }

// StoreZone is the store of zone.
func (m *Module) StoreZone() crud.Store { return m.store }

// StoreNetworkInterface is the store of network_interface.
func (m *Module) StoreNetworkInterface() crud.Store { return m.store }
