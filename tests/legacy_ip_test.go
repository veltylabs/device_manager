package tests

import (
	"strings"
	"testing"

	"github.com/veltylabs/device_manager/migrate"
)

type recordingExecer struct{ queries []string }

func (r *recordingExecer) Exec(query string, args ...any) error {
	r.queries = append(r.queries, query)
	return nil
}

// El paso legado es una sola sentencia, protegida por la existencia de device.ip:
// en una base nueva no hace nada y correrlo dos veces es seguro.
func TestLegacyDeviceIP_OneGuardedStatement(t *testing.T) {
	r := &recordingExecer{}
	if err := migrate.LegacyDeviceIP(r, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.queries) != 1 {
		t.Fatalf("queries = %d, want 1", len(r.queries))
	}
	q := r.queries[0]
	for _, want := range []string{"column_name = 'ip'", "INSERT INTO network_interface", "DROP COLUMN ip"} {
		if !strings.Contains(q, want) {
			t.Errorf("the statement must contain %q", want)
		}
	}
}
