package devicemanager

// IPLocator answers "which device has this IP" for one tenant. It has exactly
// the shape staff_manager's DeviceReader port asks for, so it satisfies it
// structurally — neither module imports the other.
type IPLocator struct {
	Devices  *Module
	TenantID string
}

func (l IPLocator) FindByIP(ip string) (string, bool) {
	d, err := l.Devices.FindByIP(l.TenantID, ip)
	if err != nil {
		return "", false
	}
	return d.Id, true
}
