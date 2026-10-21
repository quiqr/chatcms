package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"-version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "chatcmsd") {
		t.Errorf("stdout = %q, want it to name the binary", stdout.String())
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"-nope"}, &stdout, &stderr); code != 2 {
		t.Errorf("Run() = %d, want 2", code)
	}
}

func TestRunFailsWithoutDatabaseURL(t *testing.T) {
	t.Setenv("CHATCMS_DATABASE_URL", "")
	var stdout, stderr bytes.Buffer

	if code := Run(nil, &stdout, &stderr); code != 1 {
		t.Errorf("Run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "CHATCMS_DATABASE_URL") {
		t.Errorf("stderr = %q, want it to name the missing variable", stderr.String())
	}
}

func TestNewLoggerPicksHandlerByFormat(t *testing.T) {
	cases := map[string]struct {
		format   string
		terminal bool
		wantJSON bool
	}{
		"json is always json":  {"json", true, true},
		"text is always text":  {"text", false, false},
		"auto on a terminal":   {"auto", true, false},
		"auto when redirected": {"auto", false, true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			NewLogger(&buf, tc.format, tc.terminal).Info("hello", "key", "value")

			isJSON := strings.HasPrefix(strings.TrimSpace(buf.String()), "{")
			if isJSON != tc.wantJSON {
				t.Errorf("output = %q, wantJSON = %v", buf.String(), tc.wantJSON)
			}
		})
	}
}

func TestIsTerminalIsFalseForABuffer(t *testing.T) {
	if IsTerminal(&bytes.Buffer{}) {
		t.Error("IsTerminal(buffer) = true, want false")
	}
}
