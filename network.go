package devicemanager

import (
	"webtyp.com/input"
	"webtyp.com/model"
	"webtyp.com/network"
	"webtyp.com/router"
	"webtyp.com/crud"
)

// Motivos que ImportHosts entrega para una entrada descubierta que no importa.
//
// Ejemplo: motivo := devicemanager.SkipNoMAC
const (
	SkipNoMAC             = "no MAC"
	SkipNoIP              = "no IP"
	SkipRandomizedMAC     = "randomized MAC"
	SkipAlreadyRegistered = "MAC already registered"
)

// importedNamePrefix asigna nombre a un equipo importado cuyo registro en el router no tenía comentario.
const importedNamePrefix = "Imported "

// NetworkHosts adapta el inventario de un tenant para webtyp.com/network.
//
// Ejemplo: nh := devicemanager.NetworkHosts{Devices: dm, TenantID: tenantID}
type NetworkHosts struct {
	Devices  *Module
	TenantID string
}

var (
	_ network.HostSource   = NetworkHosts{}
	_ network.HostImporter = NetworkHosts{}
)

// Hosts retorna un network.Host por cada interfaz CON dirección MAC de un equipo activo.
//
// Ejemplo: hosts, err := nh.Hosts()
func (n NetworkHosts) Hosts() ([]network.Host, error) {
	devices, err := n.Devices.ListDevices(n.TenantID, DeviceFilter{ActiveOnly: true})
	if err != nil {
		return nil, err
	}
	ifaces, err := n.Devices.ListNetworkInterfaces(n.TenantID, "", "", "", 0)
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

// ImportHosts crea un equipo e interfaz por cada MAC descubierta; ver
// docs/ARCHITECTURE.md "Frontera de red" para saber qué se descarta y por qué.
//
// Ejemplo: res, err := nh.ImportHosts(discovered)
func (n NetworkHosts) ImportHosts(found []network.Discovered) (network.ImportResult, error) {
	var res network.ImportResult
	zones, err := n.Devices.ListZones(n.TenantID, "", "", 0)
	if err != nil {
		return res, err
	}

	c := crud.LoopbackActingAs(n.Devices.pub, "importer", n.Devices)
	opCreateDev := router.Operation{Module: ModelName, Entity: "device", Action: model.Create}
	opCreateNi := router.Operation{Module: ModelName, Entity: "network_interface", Action: model.Create}
	opDelDev := router.Operation{Module: ModelName, Entity: "device", Action: model.Delete}

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

		dev := &Device{Name: name, Type: DeviceTypeOther, Access: access, ZoneId: zoneID, IsActive: true}
		var errDev error
		c.Call(opCreateDev, &crud.CreateArgs{Records: []model.Model{dev}}, nil, func(e error) { errDev = e })
		if errDev != nil {
			if model.KindOf(errDev) == model.ErrorInvalid {
				skip(errDev.Error())
				continue
			}
			return res, errDev
		}

		ni := &NetworkInterface{DeviceId: dev.Id, Label: ImportedInterfaceLabel, Mac: mac, Ip: f.IP}
		var errNi error
		c.Call(opCreateNi, &crud.CreateArgs{Records: []model.Model{ni}}, nil, func(e error) { errNi = e })
		if errNi != nil {
			c.Call(opDelDev, &crud.DeleteArgs{IDs: []string{dev.Id}}, nil, func(e error) {})
			if model.KindOf(errNi) == model.ErrorInvalid {
				skip(errNi.Error())
				continue
			}
			if model.KindOf(errNi) == model.ErrorConflict {
				skip(errNi.Error())
				continue
			}
			return res, errNi
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

// dashedMAC escribe una dirección MAC canónica con guiones '-': el carácter ':' no se
// permite en el nombre de un equipo.
func dashedMAC(mac string) string {
	b := []byte(mac)
	for i := range b {
		if b[i] == ':' {
			b[i] = '-'
		}
	}
	return string(b)
}
