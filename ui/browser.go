package ui

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/svg"
)

// Browser builds this module's view for the app's registry.
func Browser(caller router.Caller, ids model.IDGenerator, tenantID string) (platformd.UIModule, error) {
	v, err := crudview.New(crudview.Config{
		ParentID:  ID,
		Presenter: devicemanager.NewView(caller),
		IDs:       ids,
	})
	if err != nil {
		return nil, err
	}
	return platformd.NewUIModule(ID, Label, svg.Icon(ID), v), nil
}
