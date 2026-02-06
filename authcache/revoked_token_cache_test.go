package authcache

import (
	"testing"
	"time"
)

func TestRevokedTokenCache(t *testing.T) {
	Add("token", time.Now().Add(1*time.Minute))
	if !IsRevoked("token") {
		t.Fatalf("expected token to be revoked")
	}

	Add("expired", time.Now().Add(-1*time.Minute))
	if IsRevoked("expired") {
		t.Fatalf("expected expired token to be not revoked")
	}
}
