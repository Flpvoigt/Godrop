package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunShowsHelpWithoutArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "GoDrop") {
		t.Fatalf("Run() stdout = %q, want help text", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run() stderr = %q, want empty", stderr.String())
	}
}

func TestRunShowsVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	oldVersion := Version
	Version = "1.2.3"
	t.Cleanup(func() { Version = oldVersion })

	code := Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}
	if got, want := stdout.String(), "godrop 1.2.3\n"; got != want {
		t.Fatalf("Run() stdout = %q, want %q", got, want)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run([]string{"desconhecido"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "comando desconhecido") {
		t.Fatalf("Run() stderr = %q, want error message", stderr.String())
	}
}

func TestRunRejectsInvalidReceivePort(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run([]string{"receive", "--port", "70000"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "porta") {
		t.Fatalf("Run() stderr = %q, want port error", stderr.String())
	}
}

func TestRunRequiresSendArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run([]string{"send"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "uso: godrop send") {
		t.Fatalf("Run() stderr = %q, want usage", stderr.String())
	}
}
