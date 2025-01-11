package crdb

import "testing"

func TestStartTestDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, cleanup := StartTestDB(t, false)
	t.Cleanup(cleanup)

	if db == nil {
		t.Errorf("db should not be nil")
	}
}
