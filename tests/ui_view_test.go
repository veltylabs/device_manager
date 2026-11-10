package tests

import (
	"webtyp.com/model"
	"webtyp.com/router"
)

type uiFakeCaller struct {
	calls []router.Operation
}

func (f *uiFakeCaller) Call(op router.Operation, args model.Encodable, out model.Decodable, done func(error)) {
	f.calls = append(f.calls, op)
	if done != nil {
		done(nil)
	}
}

func (f *uiFakeCaller) Dispatch(op router.Operation, args model.Encodable) {}
