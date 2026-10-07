# device_manager — Database Diagram

```mermaid
flowchart TD
    Z[zone<br/>id PK · tenant_id · name<br/>range_start · range_end IPv4 · updated_at]
    D[device<br/>id PK · tenant_id · name · type<br/>location · zone_id · access · is_active · updated_at]
    N[network_interface<br/>id PK · tenant_id · device_id FK<br/>label · mac · ip · updated_at]
    D -. zone_id soft reference, may be empty .-> Z
    N -->|device_id FK| D
```

- `device.type`: `computer` / `printer` / `server` / `other`.
- `device.access`: `local` / `internet_filtered` / `internet` (`webtyp.com/network` names).
- `network_interface.mac`: optional, canonical, unique per tenant when present; randomized MACs rejected.
- `network_interface.ip`: unique per tenant; auto-assigned from the device's zone when empty.
- Uniqueness and range rules are enforced in the service layer, per `tenant_id`, not as DB constraints.
