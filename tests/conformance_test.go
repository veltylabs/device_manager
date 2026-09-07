package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/router/mock"
	"webtyp.com/view"
	devicemanager "github.com/veltylabs/device_manager"
)

type fakeCaller struct {
	reply func(op string, into model.Decodable)
}

func (f *fakeCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	if f.reply != nil {
		f.reply(op, into)
	}
	if done != nil {
		done(nil)
	}
}

func (f *fakeCaller) Dispatch(op string, args model.Encodable) {}

func TestMountOperations_CreateDevice(t *testing.T) {
	m := setup(t)
	if m.ModelName() != "device_manager" {
		t.Fatalf("expected ModelName %q, got %q", "device_manager", m.ModelName())
	}

	reg := &mock.Router{}
	m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"tenant_id":"tenant-A","name":"Pc1","ip":"192.168.1.20","type":"computer","is_active":true}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+devicemanager.OpCreateDevice, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("expected success status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
	if len(ctx.ResponseBody()) == 0 {
		t.Fatal("expected a non-empty response body")
	}
}

func TestMountOperations_CreateDevice_DecodeError(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{InBody: []byte(`not valid json`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+devicemanager.OpCreateDevice, ctx)

	if ctx.Status != 400 {
		t.Fatalf("expected 400 for a malformed body, got %d", ctx.Status)
	}
}

func TestMountOperations_GetDevice_NotFound(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{InBody: []byte(`{"tenant_id":"tenant-A","id":"does-not-exist"}`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+devicemanager.OpGetDevice, ctx)

	if ctx.Status != 404 {
		t.Fatalf("expected 404 for a non-existent id, got %d", ctx.Status)
	}
}

func TestMountOperations_CreateDevice_RBACDenial(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOperations(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return false },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"tenant_id":"tenant-A","name":"Pc1","ip":"192.168.1.21","type":"computer","is_active":true}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+devicemanager.OpCreateDevice, ctx)

	if ctx.Status != 403 {
		t.Fatalf("expected 403 status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
}

func TestView_ListPopulatesItems(t *testing.T) {
	caller := &fakeCaller{
		reply: func(op string, into model.Decodable) {
			if op != devicemanager.OpListDevices {
				return
			}
			list := into.(*devicemanager.DeviceList)
			rec := list.Append().(*devicemanager.Device)
			rec.Id, rec.Name, rec.Ip = "dev_1", "Pc Recepcion", "192.168.1.30"
		},
	}
	p := devicemanager.NewView(caller)
	if err := p.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	items := p.Items()
	if len(items) != 1 || items[0].ID != "dev_1" || items[0].Label != "Pc Recepcion" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if _, ok := p.(view.Saver); !ok {
		t.Error("expected Presenter to implement view.Saver")
	}
	if _, ok := p.(view.Deleter); !ok {
		t.Error("expected Presenter to implement view.Deleter")
	}
}
