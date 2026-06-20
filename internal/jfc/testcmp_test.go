package jfc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func assertStringEqual(t testing.TB, want string, got string) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected output (-want +got):\n%s", diff)
	}
}
