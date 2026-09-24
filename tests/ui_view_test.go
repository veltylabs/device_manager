package tests

import (
	"testing"

	"github.com/veltylabs/device_manager/ui"
	"webtyp.com/dom"
	"webtyp.com/model"
)

type uiFakeCaller struct {
	calls []string
}

func (f *uiFakeCaller) Call(op string, args model.Encodable, out model.Decodable, done func(error)) {
	f.calls = append(f.calls, op)
	if done != nil {
		done(nil)
	}
}

func (f *uiFakeCaller) Dispatch(op string, args model.Encodable) {}

func TestUIView(t *testing.T) {
	fake := &uiFakeCaller{}
	ids := &mockIDGen{}

	mod, err := ui.Browser(fake, ids, "tenant-test")
	if err != nil {
		t.Fatalf("ui.Browser: %v", err)
	}

	if mod.ModelName() != ui.ID {
		t.Errorf("ModelName() = %q, want %q", mod.ModelName(), ui.ID)
	}

	if mod.Label() != ui.Label {
		t.Errorf("Label() = %q, want %q", mod.Label(), ui.Label)
	}

	v := mod.View()
	if v == nil {
		t.Fatal("View() returned nil")
	}

	if initializer, ok := v.(interface{ Init(dom.Ctx) }); ok {
		initializer.Init(nil)
	} else {
		t.Fatal("view does not implement Init method")
	}

	if len(fake.calls) == 0 {
		t.Fatal("expected calls on Init, got none")
	}

	lastCall := fake.calls[len(fake.calls)-1]
	expectedSuffix := ui.ID + ".list_devices"
	if len(lastCall) < len(expectedSuffix) || lastCall[len(lastCall)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("expected call ending with %q, got %q", expectedSuffix, lastCall)
	}
}
