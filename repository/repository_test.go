package repository

import "testing"

func TestRepositoryPackage_Skipped(t *testing.T) {
	t.Skip("repository tests require a DB driver; skipped in unit test suite")
}
