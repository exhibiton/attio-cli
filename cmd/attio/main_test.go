package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainExitCodePropagation(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name   string
		args   []string
		env    map[string]string
		wantEC int
	}{
		{
			name:   "success version",
			args:   []string{"version"},
			wantEC: 0,
		},
		{
			name:   "usage parse error",
			args:   []string{"--unknown-flag"},
			wantEC: 2,
		},
		{
			name:   "auth required",
			args:   []string{"self"},
			env:    map[string]string{"ATTIO_API_KEY": ""},
			wantEC: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec := runMainSubprocess(t, tmp, tt.args, tt.env)
			if ec != tt.wantEC {
				t.Fatalf("unexpected exit code: got %d want %d", ec, tt.wantEC)
			}
		})
	}
}

func runMainSubprocess(t *testing.T, home string, args []string, extraEnv map[string]string) int {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=TestMainSubprocessHelper")
	env := append(os.Environ(),
		"ATTIO_MAIN_SUBPROCESS=1",
		"ATTIO_MAIN_ARGS="+strings.Join(args, "\x1f"),
		"HOME="+home,
		"ATTIO_CONFIG_PATH="+filepath.Join(home, "config.json"),
		"ATTIO_BASE_URL=",
	)
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	err := cmd.Run()
	if err == nil {
		return 0
	}

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	t.Fatalf("subprocess run failed: %v", err)
	return -1
}

func TestMainSubprocessHelper(t *testing.T) {
	if os.Getenv("ATTIO_MAIN_SUBPROCESS") != "1" {
		t.Skip("helper subprocess")
	}

	argPayload := os.Getenv("ATTIO_MAIN_ARGS")
	args := []string{}
	if argPayload != "" {
		args = strings.Split(argPayload, "\x1f")
	}
	os.Args = append([]string{"attio"}, args...)

	main()
}
