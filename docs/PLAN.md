---
PLAN: "fix: detect sentinel errors without == between interfaces (no reflection in wasm)"
EXECUTOR: jules
REVIEWER: none
---

# Plan — `device_manager`: errores centinela sin `==` entre interfaces

> Master: `webtyp/docs/NO_REFLECTION_MASTER_PLAN.md` (ola 3). Doctrina: skill `api-design`.
> **Prerrequisito:** `go get webtyp.com/orm@latest` y confirmar que existe `orm.IsNotFound`. Si falta alguna, parar y reportarlo: no implementar un sustituto local.

## 1. El problema

En TinyGo, `==`, `!=` y `switch` entre valores de interfaz compilan a `runtime.interfaceEqual`, que
llama a `reflectValueEqual(reflectlite.ValueOf(x), reflectlite.ValueOf(y))`. `error` es una interfaz:
cada `err == ErrX` mete `internal/reflectlite` (~9 KB) en el binario wasm. La regla del dueño es que
el código que compila a wasm no use reflexión nunca. `errors.Is`/`errors.As` tampoco sirven: también
usan reflectlite.

## 2. La corrección — dos patrones, ninguno más

**A. Centinelas de otros paquetes** — usar su función de consulta:

| Antes | Después |
|---|---|
| `err == orm.ErrNotFound` | `orm.IsNotFound(err)` |
| `err != orm.ErrNotFound` | `!orm.IsNotFound(err)` |
| `err == storage.ErrNoRows` | `storage.IsNoRows(err)` |

**B. Centinelas propios de este paquete** — un tipo string no exportado; se afirma una vez y se
compara el valor concreto (comparación de strings, sin reflexión):

```go
// domainError is the concrete type of this package's sentinel errors. Code
// compares them by asserting this type and comparing the value: == between two
// error values compiles, under TinyGo, to runtime.interfaceEqual, which pulls
// internal/reflectlite into the wasm binary.
type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrNotFound domainError = "<texto actual>"
	// … uno por centinela, con su texto actual
)
```

- `<texto actual>`: el string exacto que devuelve hoy el centinela (`fmt.Err("a", "b")` une las
  palabras con un espacio: `"a b"`). Un test fija cada texto: los mensajes no cambian.
- Uso, por ejemplo al traducir errores a códigos:

```go
if e, ok := err.(domainError); ok {
	switch e {
	case ErrFloorInUse, ErrRoomOverlap:
		return conflict
	case ErrNotFound:
		return notFound
	}
}
if orm.IsNotFound(err) {
	return notFound
}
```

- Un `switch err { case ErrA: … }` pasa a `if e, ok := err.(domainError); ok { switch e { … } }`.
- Si un centinela propio se envuelve antes de compararlo (`fmt.Errf("…%v", ErrX)`), la comparación
  con `==` ya no funcionaba: dejarlo igual y anotarlo en el PR, no inventar otra detección.

## 3. Sitios a cambiar (inventario del 2026-10-08)

### Código de producción

- `network.go:125` — `if _, isValidation := err.(ValidationError); isValidation || err == ErrIPAlreadyExists {`
- `ops.go:84` — `if err == ErrNotFound {`
- `ops.go:106` — `} else if err == ErrIPAlreadyExists {`
- `ops.go:128` — `} else if err == ErrNotFound {`
- `ops.go:156` — `} else if err == ErrIPAlreadyExists {`
- `ops.go:158` — `} else if err == ErrNotFound {`
- `ops.go:177` — `if err == ErrNotFound {`
- `ops.go:194` — `if err == ErrNotFound {`
- `ops.go:210` — `switch err {`
- `ops.go:211` — `case ErrNotFound, ErrZoneNotFound, ErrInterfaceNotFound:`
- `ops.go:213` — `case ErrIPAlreadyExists, ErrMACAlreadyExists, ErrZoneOverlap, ErrZoneInUse, ErrZoneFull:`
- `interface.go:15` — `if err == orm.ErrNotFound {`
- `module.go:57` — `if err == orm.ErrNotFound {`
- `module.go:73` — `if err == orm.ErrNotFound {`
- `module.go:126` — `if err == ErrZoneNotFound {`
- `zone.go:15` — `if err == orm.ErrNotFound {`

### Centinelas propios de este repo (patrón B)

- `model.go:155` — `ErrNotFound          = fmt.Err("device not found")`
- `model.go:156` — `ErrIPAlreadyExists   = fmt.Err("device ip already exists")`
- `model.go:157` — `ErrMACAlreadyExists  = fmt.Err("network interface mac already exists")`
- `model.go:158` — `ErrRandomizedMAC     = fmt.Err("randomized MAC address: on the device, turn off random (private) hardware addresses for this Wi-Fi network a`
- `model.go:159` — `ErrNoZone            = fmt.Err("ip is empty and the device has no zone to assign one from")`
- `model.go:160` — `ErrIPOutsideZone     = fmt.Err("ip is outside the device's zone range")`
- `model.go:161` — `ErrZoneFull          = fmt.Err("zone has no free ip")`
- `model.go:162` — `ErrZoneOverlap       = fmt.Err("zone range overlaps another zone")`
- `model.go:163` — `ErrZoneInUse         = fmt.Err("zone is used by a device")`
- `model.go:164` — `ErrZoneNotFound      = fmt.Err("zone not found")`
- `model.go:165` — `ErrInterfaceNotFound = fmt.Err("network interface not found")`
- `model.go:166` — `ErrInvalidZoneRange  = fmt.Err("zone range must be two IPv4 addresses with start <= end")`

### Tests (se migran igual: un solo camino también en los tests)

- `tests/locator_test.go:108` — `if err != devicemanager.ErrIPAlreadyExists {`
- `tests/zone_test.go:30` — `if _, err := m.GetZone("tenant-A", z.Id); err != devicemanager.ErrZoneNotFound {`
- `tests/zone_test.go:55` — `if _, err := m.SaveZone(overlap); err != devicemanager.ErrZoneOverlap {`
- `tests/zone_test.go:70` — `if err := m.DeleteZone("tenant-A", z.Id); err != devicemanager.ErrZoneInUse {`
- `tests/zone_test.go:78` — `if _, err := m.GetZone("tenant-B", z.Id); err != devicemanager.ErrZoneNotFound {`
- `tests/zone_test.go:81` — `if err := m.DeleteZone("tenant-B", z.Id); err != devicemanager.ErrZoneNotFound {`
- `tests/zone_test.go:86` — `if _, err := m.SaveZone(hijack); err != devicemanager.ErrZoneNotFound {`
- `tests/tenant_test.go:32` — `if _, err := m.GetDevice(tenantB, created.Id); err != devicemanager.ErrNotFound {`
- `tests/tenant_test.go:40` — `if _, err := m.UpdateDevice(hijack); err != devicemanager.ErrNotFound {`
- `tests/tenant_test.go:45` — `if err := m.DeactivateDevice(tenantB, created.Id); err != devicemanager.ErrNotFound {`
- `tests/tenant_test.go:50` — `if err := m.DeleteDevice(tenantB, created.Id); err != devicemanager.ErrNotFound {`
- `tests/interface_test.go:57` — `if _, err := m.SaveNetworkInterface(iface(d, "lan b", "", "")); err != devicemanager.ErrZoneFull {`
- `tests/interface_test.go:102` — `if _, err := m.SaveNetworkInterface(iface(d, "dup", "48:F1:7F:D9:D7:B7", "")); err != devicemanager.ErrMACAlreadyExists {`
- `tests/device_test.go:70` — `if err != devicemanager.ErrIPAlreadyExists {`
- `tests/device_test.go:88` — `if err != devicemanager.ErrNotFound {`
- `tests/device_test.go:120` — `if _, err := m.GetDevice("tenant-A", d.Id); err != devicemanager.ErrNotFound {`

Si encuentras otro `==`/`!=`/`switch` entre valores de interfaz con operandos no nil que no esté en la
lista, se migra igual. `x == nil` y `x != nil` están bien.

## 4. Tests

- Todos los tests existentes siguen verdes sin cambiar su intención.
- Un test que fija el `Error()` de cada centinela propio convertido (patrón B) contra su texto anterior.
- Si el paquete traduce errores a códigos/respuestas (por ejemplo en `ops.go`), un test por rama
  cambiada: el mismo error produce el mismo código que antes.
- `gotest` verde (vet, race, tests, wasm).

## 5. Criterios de aceptación

- `grep -rnE '(==|!=) *[A-Za-z_.]*Err[A-Za-z]*' --include=*.go . | grep -v '_temp/'` → vacío.
- `grep -rn 'switch err {' --include=*.go .` → vacío.
- `grep -rn 'errors.Is\|errors.As' --include=*.go .` → vacío.
- Ningún símbolo exportado nuevo: `git diff | grep '^+func [A-Z]'`.
- `gotest` verde.

## 6. Restricciones

Las de `AGENTS.md`, más: nada de `reflect`, `unsafe`, `errors.Is`/`errors.As`, ni `==`/`!=`/`switch`
entre valores de interfaz con operandos no nil. No tocar otros repos.
