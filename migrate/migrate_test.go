package migrate_test

import (
	"testing"

	"github.com/veltylabs/device_manager/migrate"
	"webtyp.com/ddl"
	"webtyp.com/model"
)

type dummyExecer struct{ calls []string }

func (d *dummyExecer) Exec(query string, args ...any) error {
	d.calls = append(d.calls, query)
	return nil
}

type dummyCompiler struct{}

func (d *dummyCompiler) CompileDDL(stmt ddl.Stmt, m model.Model) (string, []any, error) {
	return stmt.Table, nil, nil
}

// The three tables, in foreign-key order: network_interface references device.
func TestMigrate_CreatesTablesInFKOrder(t *testing.T) {
	execer := &dummyExecer{}
	if err := migrate.Migrate(execer, &dummyCompiler{}); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	want := []string{"zone", "device", "network_interface"}
	if len(execer.calls) != len(want) {
		t.Fatalf("got Exec calls %v, want %v", execer.calls, want)
	}
	for i := range want {
		if execer.calls[i] != want[i] {
			t.Fatalf("got Exec calls %v, want %v", execer.calls, want)
		}
	}
}
