package ui

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/fmt"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/svg"
	"webtyp.com/view"
)

// Select fields whose options come from another list of this module.
const (
	fieldZoneID   = "zone_id"
	fieldDeviceID = "device_id"
)

// Browser builds the devices screen. Its zone select is filled from list_zones
// when the screen is ACTIVATED (never at creation: nothing calls the server
// until the user opens the screen).
func Browser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	v, err := crudview.New(crudview.Config{
		ParentID:  ID,
		Presenter: devicemanager.NewView(caller),
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}
	activate := onFirstActivation(v, func() {
		fillOptions(v, fieldZoneID, devicemanager.NewZoneView(caller))
	})
	return platformd.NewUIModule(ID, Label, svg.Icon(ID), v, activate), nil
}

// ZonesBrowser builds the zones screen.
func ZonesBrowser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	v, err := crudview.New(crudview.Config{
		ParentID:  ZonesID,
		Presenter: devicemanager.NewZoneView(caller),
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}
	return platformd.NewUIModule(ZonesID, ZonesLabel, svg.Icon(ID), v), nil
}

// InterfacesBrowser builds the network interfaces screen. Its device select is
// filled from list_devices on activation.
func InterfacesBrowser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	v, err := crudview.New(crudview.Config{
		ParentID:  InterfacesID,
		Presenter: devicemanager.NewNetworkInterfaceView(caller),
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}
	activate := onFirstActivation(v, func() {
		fillOptions(v, fieldDeviceID, devicemanager.NewView(caller))
	})
	return platformd.NewUIModule(InterfacesID, InterfacesLabel, svg.Icon(ID), v, activate), nil
}

// onFirstActivation returns the platformd onActivate hook: load runs once, the
// first time the screen is activated; the crudview's own lazy list load runs
// every time (platformd calls onActivate INSTEAD of the view's Activate).
func onFirstActivation(v *crudview.CrudView, load func()) func() {
	loaded := false
	return func() {
		if !loaded {
			loaded = true
			load()
		}
		v.Activate()
	}
}

// fillOptions loads source and sets its rows (ID → Label) as the options of
// v's select field. A failed load leaves the select empty; the list itself
// reports its own load errors.
func fillOptions(v *crudview.CrudView, field string, source view.Presenter) {
	source.Reload(func(err error) {
		if err != nil {
			return
		}
		items := source.Items()
		opts := make([]fmt.KeyValue, 0, len(items))
		for _, it := range items {
			opts = append(opts, fmt.KeyValue{Key: it.ID, Value: it.Label})
		}
		v.SetOptions(field, opts...)
	})
}
