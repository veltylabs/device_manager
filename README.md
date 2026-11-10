# device_manager
<img src="docs/img/badges.svg">

## ¿Por qué existe?

`device_manager` es el inventario de equipos de red de una clínica: computadores, impresoras, servidores y otros equipos, sus
interfaces de red (MAC e IP), las zonas donde se ubican y el nivel de acceso a la red de cada uno. Lo encuentras cuando
[`network_manager`](https://github.com/veltylabs/network_manager) necesita saber qué aplicar al router, o cuando el inicio de
sesión por red local pregunta qué equipo tiene una IP.

- **Sin esto:** cada módulo guarda su propia lista de equipos y las IP se repiten o se contradicen.
- **Con esto:** hay una sola fuente de verdad; la persistencia la hace [`crud`](https://github.com/webtyp/crud) y las reglas
  del negocio (IP única por inquilino, MAC no aleatoria, rangos de zona sin solapar) viven en los archivos `rule_*.go`
  ([arquitectura](docs/ARCHITECTURE.md)).

## Antes de comenzar

- Go instalado y acceso a los módulos `webtyp.com/*` (`GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`).
- Una base de datos conectada como `*orm.DB`; las tablas las crea [`migrate.Migrate`](migrate/).
- Un generador de ids (`model.IDGenerator`) y el id del inquilino (`TenantID`): un módulo sirve a un solo inquilino.

## Guía paso a paso

### Paso 1: crear el módulo

**Objetivo:** conectar el inventario a una base en memoria.

```go
package main

import (
	devicemanager "github.com/veltylabs/device_manager"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

func main() {
	ids, err := unixid.NewUnixID()
	if err != nil {
		fmt.Println(err)
		return
	}
	mod, err := devicemanager.New(orm.New(mem.New()), devicemanager.Deps{IDs: ids, TenantID: "acme"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mod.ModelName())
}
```

**Resultado esperado:** imprime `device_manager`.

### Paso 2: registrar un equipo

**Objetivo:** crear un equipo y comprobar que una IP repetida se rechaza.

`CreateDevice`, `SaveNetworkInterface` y los demás métodos de escritura son envoltorios: ejecutan la operación del propio
módulo (`device.create`, `network_interface.create`…) por su `Loopback`, de modo que las reglas de `rule_*.go` son el único
camino que escribe. Los ejemplos completos están en [tests/device_test.go](tests/device_test.go) y
[tests/interface_test.go](tests/interface_test.go).

**Resultado esperado:** el equipo queda guardado; una segunda interfaz con la misma IP recibe un conflicto.

### Paso 3: ver el módulo en pantalla

**Objetivo:** abrir la demo con datos en memoria. La demo está en [seed/demo.go](seed/demo.go) y se abre con `webtyp` en la raíz del módulo.

**Resultado esperado:** listas de equipos, zonas e interfaces con búsqueda.

## Operaciones: sin y con `crud`

| Antes (manejadores a mano) | Después (declaración + reglas) |
|---|---|
| `list_devices`, `get_device` | `device.read` |
| `create_device`, `update_device`, `upsert_device`, `deactivate_device` | `device.create`, `device.update` (`deactivate` = `update` de `is_active`) |
| `delete_device` | `device.delete` (borra antes sus interfaces, en la misma transacción) |
| `list_zones`, `save_zone`, `delete_zone` | `zone.read`, `zone.create`/`zone.update`, `zone.delete` |
| `list_network_interfaces`, `save_network_interface`, `delete_network_interface` | `network_interface.read`, `.create`/`.update`, `.delete` |

El permiso de cada una es su recurso: `device`, `zone` y `network_interface`.

## Mapa de documentación

| Orden | Documento | Qué pregunta responde y cuándo leerlo |
|---|---|---|
| 1 | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | ¿Qué entidades hay y quién decide qué? Léelo antes de tocar una regla. |
| 2 | [docs/TESTS.md](docs/TESTS.md) | ¿Qué comportamiento está probado? Léelo antes de cambiar una regla. |
| 3 | [docs/diagrams/database.md](docs/diagrams/database.md) | ¿Cómo se relacionan las tablas? Léelo antes de cambiar el esquema. |
