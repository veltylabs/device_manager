# Bugs de UI y mejoras — `device_manager`

> Auditoría en vivo sobre `#device_manager`, `#device_manager_zones`, `#device_manager_interfaces` (2026-11-19).

## Bugs encontrados

| # | Sev | Pantalla | Qué se ve en pantalla | Causa raíz verificada |
|---|---|---|---|---|
| DM-1 | 🔴 | `#device_manager_interfaces` | **Las 24 filas de la lista dicen exactamente `lan` sin indicar a qué equipo pertenece cada interfaz.** Es imposible distinguir qué PC corresponde a qué IP (`172.0.0.81`, `172.0.0.82`, etc.) sin abrir las 24 filas una por una. | `NetworkInterfaceModel` (`device_manager/definition.go:84-87`) marca `label` (`"lan"`) como `ItemLabel` e `ip` como `ItemDescription`, pero omite el nombre del `device_id` en la fila. |
| DM-2 | 🟡 | `#device_manager` | **Botones de radio `Tipo` y `Acceso` tienen un hueco vacío gigante debajo del texto** dentro de cada chip (`56px` de alto con texto arriba). | Bug compartido del framework `webtyp/form/css.go:105` (`PartRadioOption` con `display: block` sin centrado flex — F1 en el master). |
| DM-3 | 🟡 | Menú | **Aporta 3 enlaces independientes al menú raíz (`Equipos`, `Zonas`, `Interfaces de red`)** sin agrupación ni jerarquía. | `device_manager/ui` no agrupa subpantallas bajo un único módulo padre (viola U1 de `APP_USABILITY_MASTER_PLAN` y «una pantalla por tarea» de `ui-consistency`). |
| DM-4 | 🟡 | `#device_manager_zones` | **`Inicio de rango` y `Fin de rango` muestran exactamente el mismo ejemplo (`Ejemplo: 192.168.1.1`)** y la lista `0 / 0` carece de guía de estado vacío. | Ambos campos usan `input.IP()` (`device_manager/definition.go:67-68`) con placeholder genérico idéntico. |
| DM-5 | ⚪ | `#device_manager` | **Badge truncado en fila (`Oficina 2 · ...`)** cuando se combina ubicación y acceso. | Ancho fijo del badge en `targetlist` al concatenar varios campos en `ItemDescription`. |

## Sugerencias de mejora justificadas

1. **Integrar «Interfaces de red» dentro de la ficha de «Equipos» (una pantalla por tarea):** mostrar y editar las tarjetas de red (etiqueta, MAC, IP) directamente debajo o junto al equipo seleccionado, o al menos anteponer el nombre del equipo en el título de la fila (`pc farmacia · lan`). *Justificación:* en la clínica nadie administra interfaces sueltas sin saber de qué computador son; elimina 1 enlace del menú y resuelve DM-1 de raíz.
2. **Agrupar `Equipos` y `Zonas` bajo un solo grupo padre «Equipos»:** *Justificación:* libera espacio en la barra lateral de `mjosefa-cms` y cumple U1.
3. **Diferenciar placeholders de rango en `Zonas` (`192.168.1.10` / `192.168.1.254`) y agregar mensaje guía cuando `0 / 0`:** *Justificación:* explica al administrador qué es una zona antes de que cree la primera.
