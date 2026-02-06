package database

import "testing"

func TestDatabasePackage_Skipped(t *testing.T) {
	t.Skip("database tests require a live DB or sqlite driver; skipped in unit test suite")
}
