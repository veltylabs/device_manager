package devicemanager

import (
	"webtyp.com/input"
	"webtyp.com/network"
)

// Reasons ImportHosts gives for a discovered entry it does not import.
const (
	SkipNoMAC             = "no MAC"
	SkipNoIP              = "no IP"
	SkipRandomizedMAC     = "randomized MAC"
	SkipAlreadyRegistered = "MAC already registered"
)

// importedNamePrefix names an imported device whose router entry had no comment.
const importedNamePrefix = "Imported "

// NetworkHosts adapts the inventory of one tenant to webtyp.com/network.
type NetworkHosts struct {
	Devices  *Module
	TenantID string
}

var (
	_ network.HostSource   = NetworkHosts{}
	_ network.HostImporter = NetworkHosts{}
)

// Hosts returns one network.Host per interface WITH a MAC of an active device.
func (n NetworkHosts) Hosts() ([]network.Host, error) {
	devices, err := n.Devices.ListDevices(n.TenantID, DeviceFilter{ActiveOnly: true})
	if err != nil {
		return nil, err
	}
	ifaces, err := n.Devices.ListNetworkInterfaces(n.TenantID, "")
	if err != nil {
		return nil, err
	}
	var hosts []network.Host
	for _, ni := range ifaces {
		if ni.Mac == "" {
			continue
		}
		for _, d := range devices {
			if d.Id != ni.DeviceId {
				continue
			}
			access, err := network.ParseAccess(d.Access)
			if err != nil {
				return nil, err
			}
			hosts = append(hosts, network.Host{
				Name:   d.Name + " (" + ni.Label + ")",
				MAC:    ni.Mac,
				IP:     ni.Ip,
				Access: access,
			})
			break
		}
	}
	return hosts, nil
}

// ImportHosts creates one device + interface per discovered MAC; see
// docs/ARCHITECTURE.md "Network boundary" for what is skipped and why.
func (n NetworkHosts) ImportHosts(found []network.Discovered) (network.ImportResult, error) {
	var res network.ImportResult
	zones, err := n.Devices.ListZones(n.TenantID)
	if err != nil {
		return res, err
	}
	for _, f := range found {
		skip := func(reason string) { res.Skipped = append(res.Skipped, network.Skipped{Found: f, Reason: reason}) }
		if f.MAC == "" {
			skip(SkipNoMAC)
			continue
		}
		if f.IP == "" {
			skip(SkipNoIP)
			continue
		}
		mac := input.CanonicalMAC(f.MAC)
		if input.IsLocallyAdministeredMAC(mac) {
			skip(SkipRandomizedMAC)
			continue
		}
		registered, err := n.macRegistered(mac)
		if err != nil {
			return res, err
		}
		if registered {
			skip(SkipAlreadyRegistered)
			continue
		}

		name := f.Name
		if name == "" {
			name = importedNamePrefix + dashedMAC(mac)
		}
		access := network.AccessLocalName
		if f.Internet {
			access = network.AccessInternetName
		}
		zoneID := ""
		if z, ok := zoneContaining(zones, f.IP); ok {
			zoneID = z.Id
		}
		dev, err := n.Devices.CreateDevice(Device{TenantId: n.TenantID, Name: name, Type: DeviceTypeOther,
			Access: access, ZoneId: zoneID, IsActive: true})
		if err != nil {
			if _, isValidation := err.(ValidationError); isValidation {
				skip(err.Error())
				continue
			}
			return res, err
		}
		_, err = n.Devices.SaveNetworkInterface(NetworkInterface{TenantId: n.TenantID, DeviceId: dev.Id,
			Label: ImportedInterfaceLabel, Mac: mac, Ip: f.IP})
		if err != nil {
			// The device must not stay without its interface.
			if delErr := n.Devices.DeleteDevice(n.TenantID, dev.Id); delErr != nil {
				return res, delErr
			}
			if _, isValidation := err.(ValidationError); isValidation || err == ErrIPAlreadyExists {
				skip(err.Error())
				continue
			}
			return res, err
		}
		res.Created++
	}
	return res, nil
}

func (n NetworkHosts) macRegistered(mac string) (bool, error) {
	var ni NetworkInterface
	rows, err := ReadAllNetworkInterface(n.Devices.db.Query(&ni).
		Where(NetworkInterface_.Mac).Eq(mac).
		Where(NetworkInterface_.TenantId).Eq(n.TenantID).Limit(1))
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

// dashedMAC writes a canonical MAC with '-' separators: ':' is not allowed in
// a device name.
func dashedMAC(mac string) string {
	b := []byte(mac)
	for i := range b {
		if b[i] == ':' {
			b[i] = '-'
		}
	}
	return string(b)
}
