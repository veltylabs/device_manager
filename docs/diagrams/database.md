# device_manager — Diagrama de Base de Datos

```mermaid
flowchart TD
    Z["zone<br/>id PK · tenant_id · name<br/>range_start · range_end IPv4 · updated_at"]
    D["device<br/>id PK · tenant_id · name · type<br/>location · zone_id · access · is_active · updated_at"]
    N["network_interface<br/>id PK · tenant_id · device_id FK<br/>label · mac · ip · updated_at"]
    D -. "zone_id referencia blanda, puede ser vacía" .-> Z
    N -->|"device_id FK"| D
```

- `device.type`: `computer` / `printer` / `server` / `other`.
- `device.access`: `local` / `internet_filtered` / `internet` (nombres de `webtyp.com/network`).
- `network_interface.mac`: opcional, canónica, única por tenant cuando está presente; se rechazan MACs aleatorias.
- `network_interface.ip`: única por tenant; autoasignada desde la zona del equipo cuando está vacía.
- Las reglas de unicidad y rangos se aplican en la capa de servicio, por `tenant_id`, no como restricciones de base de datos.
