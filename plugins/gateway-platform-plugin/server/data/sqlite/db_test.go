package sqlite

import "testing"

func TestInitSchemaCreatesCoreTables(t *testing.T) {
	db, err := OpenTestDB(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatal(err)
	}

	for _, table := range []string{"routes", "keys", "route_rewrites"} {
		if !TableExists(t, db, table) {
			t.Fatalf("expected table %s to exist", table)
		}
	}
}
