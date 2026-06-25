package main

import "testing"

func TestParsePodPrefix(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantPodID string
		wantRest  string
	}{
		{
			name:      "extracts pod name and strips prefix from json line",
			line:      `[pod/my-service-abc123-xyz/main] {"level":"info","msg":"hello"}`,
			wantPodID: "my-service-abc123-xyz",
			wantRest:  `{"level":"info","msg":"hello"}`,
		},
		{
			name:      "returns empty pod id and untouched line when no prefix",
			line:      `{"level":"info","msg":"hello"}`,
			wantPodID: "",
			wantRest:  `{"level":"info","msg":"hello"}`,
		},
		{
			name:      "preserves trailing newline in rest",
			line:      "[pod/svc-1/main] plain text line\n",
			wantPodID: "svc-1",
			wantRest:  "plain text line\n",
		},
		{
			name:      "no match when prefix has too few segments",
			line:      "[pod/only-one] not a real prefix",
			wantPodID: "",
			wantRest:  "[pod/only-one] not a real prefix",
		},
		{
			name:      "no match when bracket is unclosed",
			line:      "[pod/svc/main {\"level\":\"info\"}",
			wantPodID: "",
			wantRest:  "[pod/svc/main {\"level\":\"info\"}",
		},
		{
			name:      "no match for unrelated bracketed content",
			line:      "[some other thing] message",
			wantPodID: "",
			wantRest:  "[some other thing] message",
		},
		{
			name:      "handles empty line",
			line:      "",
			wantPodID: "",
			wantRest:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPodID, gotRest := parsePodPrefix([]byte(tt.line))

			if gotPodID != tt.wantPodID {
				t.Errorf("parsePodPrefix(%q) podID = %q, want %q", tt.line, gotPodID, tt.wantPodID)
			}
			if string(gotRest) != tt.wantRest {
				t.Errorf("parsePodPrefix(%q) rest = %q, want %q", tt.line, string(gotRest), tt.wantRest)
			}
		})
	}
}
