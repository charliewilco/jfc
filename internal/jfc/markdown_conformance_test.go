package jfc

import "testing"

func TestFormatMarkdownPreservesRenderedConformanceCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{name: "safe_blank_line_spaces", input: "# Title\n   \nParagraph.\n"},
		{name: "hard_break_trailing_spaces", input: "line with hard break  \nnext line\n"},
		{name: "fenced_code_block", input: "```go\nfmt.Println(\"x\")\n```\n"},
		{name: "tilde_fence_with_indent", input: "  ~~~json\n  {\"x\":1}\n  ~~~\n"},
		{name: "indented_code_block", input: "    alpha\n    \n    beta\n"},
		{name: "nested_list", input: "- alpha\n  - beta\n    continuation\n"},
		{name: "blockquote_blank_line", input: "> alpha\n>\n> beta\n"},
		{name: "table_like_text", input: "| a | b |\n| - | - |\n| 1 | 2 |\n"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output, err := formatMarkdown([]byte(tc.input), DefaultConfig())
			if err != nil {
				t.Fatalf("formatMarkdown returned error: %v", err)
			}

			idempotent, err := formatMarkdown(output, DefaultConfig())
			if err != nil {
				t.Fatalf("formatMarkdown rejected its own output: %v\noutput:\n%s", err, output)
			}
			assertStringEqual(t, string(output), string(idempotent))
			assertMarkdownHTMLSemanticallyEqual(t, []byte(tc.input), output)
		})
	}
}
