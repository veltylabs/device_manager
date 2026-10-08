package tests

func isErr(err error, target error) bool {
    if err == nil {
        return target == nil
    }
    if target == nil {
        return false
    }
    return err.Error() == target.Error()
}
