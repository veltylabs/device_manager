package migrate

import (
	"webtyp.com/ddl"

	devicemanager "github.com/veltylabs/device_manager"
)

// Migrate reconcilia el esquema de base de datos de device_manager: Zone, Device,
// NetworkInterface.
//
// No es invocado por New y vive deliberadamente en su propio paquete: nada en la ruta
// de compilación WASM de una aplicación consumidora importa este paquete, evitando que
// webtyp.com/ddl ingrese en dicho grafo de dependencias.
//
// conn es un ddl.Execer, no un *orm.DB, por lo que un ejecutor DDL de despliegue lo satisface.
// RawConn() de *orm.DB también lo satisface para pruebas y entornos locales.
//
// Ejemplo: err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	// Sync y no CreateTable: crea la tabla que falta y además agrega las columnas que un modelo
	// ganó, así un campo nuevo llega a una base ya instalada. Flujo: https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md
	return ddl.New(conn, ddlCompiler).Sync(
		&devicemanager.Zone{},
		&devicemanager.Device{},
		&devicemanager.NetworkInterface{},
	)
}
