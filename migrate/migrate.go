package migrate

import (
	"webtyp.com/ddl"

	devicemanager "github.com/veltylabs/device_manager"
)

// Migrate reconciles the database schema device_manager owns: Zone, Device,
// NetworkInterface.
//
// It is deliberately NOT called by New, and deliberately lives in its own
// package rather than a new file in the root package: nothing on a
// consuming app's WASM build path (its view.go, which imports the root
// devicemanager package for devicemanager.NewView) ever imports
// "github.com/veltylabs/device_manager/migrate" — so webtyp.com/ddl never
// enters that build graph, regardless of build tags on the consumer's side.
//
// conn is a ddl.Execer, not an *orm.DB, so a deploy-time transport that can
// only execute DDL satisfies it. An *orm.DB's RawConn() also satisfies it,
// for local/test callers:
//
//	conn, _ := postgres.Open(dsn)
//	compiler, _ := conn.(ddl.Compiler)
//	err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	d := ddl.New(conn, ddlCompiler)
	// FK order: network_interface.device_id references device.id.
	if err := d.CreateTable(&devicemanager.Zone{}); err != nil {
		return err
	}
	if err := d.CreateTable(&devicemanager.Device{}); err != nil {
		return err
	}
	return d.CreateTable(&devicemanager.NetworkInterface{})
}
