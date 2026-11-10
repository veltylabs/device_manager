package tests

import (
	"strings"

	"webtyp.com/model"
)

func isErr(err error, target error) bool {
	if err == nil {
		return target == nil
	}
	if target == nil {
		return false
	}
	// a loopback answer prefixes the domain text with "loopback: <status> ": compare the kind and the end of the text
	return model.KindOf(err) == model.KindOf(target) && strings.HasSuffix(err.Error(), target.Error())
}
