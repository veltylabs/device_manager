package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
	devicemanager "github.com/veltylabs/device_manager"
)

// TestOpListDevices_FallsBackToModuleTenant reproduces a real bug found while
// manually testing a downstream app's "Equipos" screen: the list showed
// "0 / 0" even though a real device row existed for the app's tenant.
//
// Root cause: webtyp.com/view's callerLister.list() — the code every
// crudview-backed list runs on Reload() — always calls the List op with NO
// args (see router.Caller's own doc: this is correct, by design, since a
// List op is written to tolerate an absent/empty args object). opListDevices
// therefore receives ListDevicesArgs{TenantId: ""} on every real page load,
// and ListDevices("", filter) legitimately (and silently — no error) finds
// zero rows, because no real device has tenant_id == "".
//
// Every OTHER caller of this op (an app's own explicit filtered query) can
// still pass a real tenant_id and gets a correct, scoped result — this bug
// affects ONLY the args-less path the UI actually uses on every page load.
//
// github.com/veltylabs/staff_manager already solves this exact problem:
// Deps carries a TenantID, New requires it, and opListStaff falls back to it
// when the caller's args are empty. device_manager has no such field or
// fallback. This test seeds one device under the module's own configured
// tenant and invokes OpListDevices with an EMPTY body — exactly what
// callerLister.list() sends — expecting that seeded row back.
//
// It does not compile today (devicemanager.Deps has no TenantID field) —
// that is the intended red state; adding the field and the fallback is
// stage 1 of docs/PLAN.md.
func TestOpListDevices_FallsBackToModuleTenant(t *testing.T) {
	const moduleTenant = "mjosefa-cms"

	db := orm.New(mem.New())
	m, err := devicemanager.New(db, devicemanager.Deps{IDs: &mockIDGen{}, TenantID: moduleTenant})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	seeded, err := m.CreateDevice(devicemanager.Device{
		TenantId: moduleTenant,
		Name:     "Front Desk PC",
		Ip:       "192.168.1.50",
		Type:     devicemanager.DeviceTypeComputer,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}

	reg := &mock.Router{}
	m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	// The exact wire shape view.NewCallerLister sends: no args at all.
	ctx := &mock.Context{InBody: []byte(`{}`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+devicemanager.OpListDevices, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("list_devices with empty args: status = %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
	body := string(ctx.ResponseBody())
	if !contains(body, seeded.Id) {
		t.Fatalf("list_devices with empty args did not return the seeded device (id %q) for the "+
			"module's own tenant %q — got: %s — this reproduces the \"0 / 0\" bug the real UI shows "+
			"on every page load, since it never sends a tenant_id", seeded.Id, moduleTenant, body)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
