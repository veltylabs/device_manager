package migrate

import "webtyp.com/ddl"

// LegacyDeviceIPStep nombra el paso que actualiza una tabla device creada antes de
// device_manager v0.2.0.
//
// Ejemplo: ddl.Step{Name: migrate.LegacyDeviceIPStep, Run: migrate.LegacyDeviceIP}
const LegacyDeviceIPStep = "device_manager: legacy device.ip"

// legacyDeviceIPSQL agrega zone_id/access, mueve cada device.ip a una fila "lan" de
// network_interface (sin MAC) y luego borra device.ip. No hace nada cuando ip ya no
// existe, así que correrlo dos veces es seguro. Solo Postgres: lee information_schema.
const legacyDeviceIPSQL = `DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name = 'device' AND column_name = 'ip') THEN
    ALTER TABLE device ADD COLUMN IF NOT EXISTS zone_id TEXT;
    ALTER TABLE device ADD COLUMN IF NOT EXISTS access TEXT NOT NULL DEFAULT 'local';
    INSERT INTO network_interface (id, tenant_id, device_id, label, ip, updated_at)
      SELECT d.id || '-lan', d.tenant_id, d.id, 'lan', d.ip, 0
      FROM device d
      WHERE d.ip IS NOT NULL AND d.ip <> ''
        AND NOT EXISTS (SELECT 1 FROM network_interface n WHERE n.device_id = d.id);
    ALTER TABLE device DROP COLUMN ip;
  END IF;
END $$;`

// LegacyDeviceIP actualiza una base Postgres cuya tabla device es anterior a
// device_manager v0.2.0. Corre después de Migrate: necesita network_interface.
//
// Ejemplo: if err := migrate.LegacyDeviceIP(conn, compiler); err != nil { return err }
func LegacyDeviceIP(conn ddl.Execer, _ ddl.Compiler) error {
	_, err := conn.Exec(legacyDeviceIPSQL)
	return err
}
