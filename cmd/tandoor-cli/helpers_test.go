// cmd/tandoor-cli/helpers_test.go

package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestApplyJQ(t *testing.T) {
	ctx := context.Background()
	data := []byte(`{"id": 1, "name": "Test", "items": [{"a": 1}, {"a": 2}]}`)

	tests := []struct {
		name    string
		filter  string
		want    string
		wantErr string
	}{
		{
			name:   "empty filter returns input unchanged",
			filter: "",
			want:   string(data),
		},
		{
			name:   "identity",
			filter: ".",
			want:   `{"id":1,"name":"Test","items":[{"a":1},{"a":2}]}`,
		},
		{
			name:   "field extraction",
			filter: ".name",
			want:   `"Test"`,
		},
		{
			name:   "multiple outputs joined by newline",
			filter: ".items[].a",
			want:   "1\n2",
		},
		{
			name:   "constructs objects",
			filter: "{id: .id, name: .name}",
			want:   `{"id":1,"name":"Test"}`,
		},
		{
			name:    "invalid filter",
			filter:  ".foo & .bar",
			wantErr: "jq filter failed",
		},
		{
			name:    "runtime error",
			filter:  `error("boom")`,
			wantErr: "jq filter failed",
		},
		{
			name:   "no matches produces empty output",
			filter: ".items[] | select(.a > 100)",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyJQ(ctx, tt.filter, data)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (output: %s)", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Each output line is one compact JSON value (matching
			// `jq -c` behavior), so compare line by line.
			gotLines := splitJSONLines(t, got)
			if tt.want == "" {
				if len(gotLines) != 0 {
					t.Fatalf("expected empty output, got %q", got)
				}
				return
			}
			wantLines := splitJSONLines(t, []byte(tt.want))
			if len(gotLines) != len(wantLines) {
				t.Fatalf("got %d output values, want %d:\n%s", len(gotLines), len(wantLines), got)
			}
			for i := range wantLines {
				gotJSON, _ := json.Marshal(normalizeJSON(t, gotLines[i]))
				wantJSON, _ := json.Marshal(normalizeJSON(t, wantLines[i]))
				if string(gotJSON) != string(wantJSON) {
					t.Fatalf("value %d: got %s, want %s", i, gotJSON, wantJSON)
				}
			}
		})
	}
}

func TestApplyJQPreservesLargeIntegers(t *testing.T) {
	ctx := context.Background()
	data := []byte(`{"big": 4722366482869645213696}`)
	got, err := applyJQ(ctx, ".big", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "4722366482869645213696\n" {
		t.Fatalf("large integer mangled: %s", got)
	}
}

func splitJSONLines(t *testing.T, data []byte) [][]byte {
	t.Helper()
	var lines [][]byte
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, []byte(line))
	}
	return lines
}

func normalizeJSON(t *testing.T, data []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}
	return v
}
