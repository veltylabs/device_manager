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
	d := ddl.New(conn, ddlCompiler)
	// Orden de dependencias FK: network_interface.device_id referencia a device.id.
	if err := d.CreateTable(&devicemanager.Zone{}); err != nil {
		return err
	}
	if err := d.CreateTable(&devicemanager.Device{}); err != nil {
		return err
	}
	return d.CreateTable(&devicemanager.NetworkInterface{})
}
