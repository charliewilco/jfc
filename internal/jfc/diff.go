package jfc

import (
	"fmt"
	"strings"
)

const diffContextLines = 3

type diffOpKind int

const (
	diffEqual diffOpKind = iota
	diffDelete
	diffInsert
)

type diffOp struct {
	kind diffOpKind
	line string
}

type diffLine struct {
	oldLine int
	newLine int
	op      diffOp
}

type diffHunk struct {
	lines    []diffLine
	oldStart int
	oldCount int
	newStart int
	newCount int
}

func unifiedDiff(oldName string, newName string, oldContent []byte, newContent []byte) string {
	oldLines := splitDiffLines(string(oldContent))
	newLines := splitDiffLines(string(newContent))
	ops := diffLineOps(oldLines.lines, newLines.lines)
	hunks := diffHunks(ops, diffContextLines)
	if len(hunks) == 0 {
		return ""
	}

	var output strings.Builder
	fmt.Fprintf(&output, "--- %s\n", oldName)
	fmt.Fprintf(&output, "+++ %s\n", newName)
	for _, hunk := range hunks {
		fmt.Fprintf(&output, "@@ -%d,%d +%d,%d @@\n", hunk.oldStart, hunk.oldCount, hunk.newStart, hunk.newCount)
		for _, line := range hunk.lines {
			switch line.op.kind {
			case diffEqual:
				output.WriteByte(' ')
			case diffDelete:
				output.WriteByte('-')
			case diffInsert:
				output.WriteByte('+')
			}
			output.WriteString(line.op.line)
			output.WriteByte('\n')
			if line.op.kind == diffDelete && !oldLines.trailingNewline && line.oldLine == len(oldLines.lines) {
				output.WriteString("\\ No newline at end of file\n")
			}
			if line.op.kind == diffInsert && !newLines.trailingNewline && line.newLine == len(newLines.lines) {
				output.WriteString("\\ No newline at end of file\n")
			}
		}
	}
	return output.String()
}

type splitDiffResult struct {
	lines           []string
	trailingNewline bool
}

func splitDiffLines(content string) splitDiffResult {
	if content == "" {
		return splitDiffResult{}
	}

	trailingNewline := strings.HasSuffix(content, "\n")
	content = strings.TrimSuffix(content, "\n")
	return splitDiffResult{
		lines:           strings.Split(content, "\n"),
		trailingNewline: trailingNewline,
	}
}

func diffLineOps(oldLines []string, newLines []string) []diffOp {
	lcs := make([][]int, len(oldLines)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(newLines)+1)
	}

	for i := len(oldLines) - 1; i >= 0; i-- {
		for j := len(newLines) - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	ops := make([]diffOp, 0, len(oldLines)+len(newLines))
	for i, j := 0, 0; i < len(oldLines) || j < len(newLines); {
		switch {
		case i < len(oldLines) && j < len(newLines) && oldLines[i] == newLines[j]:
			ops = append(ops, diffOp{kind: diffEqual, line: oldLines[i]})
			i++
			j++
		case j < len(newLines) && (i == len(oldLines) || lcs[i][j+1] > lcs[i+1][j]):
			ops = append(ops, diffOp{kind: diffInsert, line: newLines[j]})
			j++
		default:
			ops = append(ops, diffOp{kind: diffDelete, line: oldLines[i]})
			i++
		}
	}
	return ops
}

func diffHunks(ops []diffOp, context int) []diffHunk {
	lines := annotateDiffLines(ops)
	var hunks []diffHunk

	for i := 0; i < len(lines); {
		for i < len(lines) && lines[i].op.kind == diffEqual {
			i++
		}
		if i >= len(lines) {
			break
		}

		start := max(0, i-context)
		end := i
		lastChange := i
		for end < len(lines) {
			if lines[end].op.kind != diffEqual {
				lastChange = end
			}
			if end-lastChange > context {
				break
			}
			end++
		}
		if end-lastChange > context {
			end -= end - lastChange - context
		}

		hunk := makeDiffHunk(lines[start:end])
		hunks = append(hunks, hunk)
		i = end
	}

	return hunks
}

func annotateDiffLines(ops []diffOp) []diffLine {
	oldLine := 1
	newLine := 1
	lines := make([]diffLine, 0, len(ops))
	for _, op := range ops {
		line := diffLine{op: op}
		switch op.kind {
		case diffEqual:
			line.oldLine = oldLine
			line.newLine = newLine
			oldLine++
			newLine++
		case diffDelete:
			line.oldLine = oldLine
			oldLine++
		case diffInsert:
			line.newLine = newLine
			newLine++
		}
		lines = append(lines, line)
	}
	return lines
}

func makeDiffHunk(lines []diffLine) diffHunk {
	hunk := diffHunk{lines: lines}
	for _, line := range lines {
		if line.op.kind != diffInsert {
			if hunk.oldStart == 0 {
				hunk.oldStart = line.oldLine
			}
			hunk.oldCount++
		}
		if line.op.kind != diffDelete {
			if hunk.newStart == 0 {
				hunk.newStart = line.newLine
			}
			hunk.newCount++
		}
	}
	if hunk.oldStart == 0 {
		hunk.oldStart = lines[0].oldLine
	}
	if hunk.newStart == 0 {
		hunk.newStart = lines[0].newLine
	}
	return hunk
}
