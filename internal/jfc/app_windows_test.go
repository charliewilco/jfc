//go:build windows

package jfc

import (
	"path/filepath"
	"testing"
)

func TestRecursiveGlobRootPreservesWindowsDriveRoot(t *testing.T) {
	t.Parallel()

	if got := recursiveGlobRoot(`C:\**\*.json`); got != `C:\` {
		t.Fatalf(`recursiveGlobRoot drive root = %q, want C:\`, got)
	}
	if got := recursiveGlobRoot(`C:\foo*\**\*.json`); got != `C:\` {
		t.Fatalf(`recursiveGlobRoot glob after drive root = %q, want C:\`, got)
	}
	if got := recursiveGlobRoot(`C:\repo\**\*.json`); got != filepath.Clean(`C:\repo`) {
		t.Fatalf(`recursiveGlobRoot fixed directory = %q, want C:\repo`, got)
	}
}
