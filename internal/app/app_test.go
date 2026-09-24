package app

import "testing"

func TestResolveVersionUsesLinkedVersion(t *testing.T) {
	if got := resolveVersion("v0.1.3"); got != "v0.1.3" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v0.1.3")
	}
}

func TestResolveVersionKeepsDevelopmentVersion(t *testing.T) {
	if got := resolveVersion("dev"); got != "dev" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "dev")
	}
}
