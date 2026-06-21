package jfc

import "testing"

func TestDiffMatrixTooLarge(t *testing.T) {
	t.Parallel()

	if diffMatrixTooLarge(100, 100) {
		t.Fatalf("expected small diff matrix to be allowed")
	}
	if !diffMatrixTooLarge(maxDiffMatrixCells, 2) {
		t.Fatalf("expected oversized diff matrix to be rejected")
	}
}

func TestDiffLineOpsKeepsCommonContextAroundSmallChanges(t *testing.T) {
	t.Parallel()

	ops := diffLineOps([]string{"a", "b", "c"}, []string{"a", "x", "c"})
	kinds := make([]diffOpKind, 0, len(ops))
	for _, op := range ops {
		kinds = append(kinds, op.kind)
	}

	expected := []diffOpKind{diffEqual, diffDelete, diffInsert, diffEqual}
	if len(kinds) != len(expected) {
		t.Fatalf("expected %d ops, got %d: %+v", len(expected), len(kinds), ops)
	}
	for i := range expected {
		if kinds[i] != expected[i] {
			t.Fatalf("op %d kind = %v, want %v; ops = %+v", i, kinds[i], expected[i], ops)
		}
	}
}

func TestBoundedDiffLineOpsFallsBackToReplacement(t *testing.T) {
	t.Parallel()

	oldLines := make([]string, maxDiffMatrixCells/3+1)
	newLines := []string{"new-a", "new-b"}
	ops := boundedDiffLineOps(oldLines, newLines)

	if len(ops) != len(oldLines)+len(newLines) {
		t.Fatalf("expected replacement ops, got %d ops", len(ops))
	}
	if ops[0].kind != diffDelete {
		t.Fatalf("expected fallback to start with deletions, got %+v", ops[0])
	}
	if ops[len(ops)-1].kind != diffInsert || ops[len(ops)-1].line != "new-b" {
		t.Fatalf("expected fallback to end with insertions, got %+v", ops[len(ops)-1])
	}
}
