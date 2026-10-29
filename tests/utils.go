package tests

import "webtyp.com/model"

func isErr(err error, target error) bool {
    if err == nil {
        return target == nil
    }
    if target == nil {
        return false
    }
    return model.KindOf(err) == model.KindOf(target) && err.Error() == target.Error()
}
