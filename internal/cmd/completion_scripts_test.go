package cmd

import (
	"strings"
	"testing"
)

func TestCompletionScriptsAllShells(t *testing.T) {
	tests := []struct {
		shell      string
		wantMarker string
	}{
		{shell: "bash", wantMarker: "_attio_complete"},
		{shell: "zsh", wantMarker: "#compdef attio"},
		{shell: "fish", wantMarker: "complete -c attio"},
		{shell: "powershell", wantMarker: "Register-ArgumentCompleter"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			script, err := completionScript(tt.shell)
			if err != nil {
				t.Fatalf("completionScript(%s): %v", tt.shell, err)
			}
			if !strings.Contains(script, tt.wantMarker) {
				t.Fatalf("expected marker %q in %s script", tt.wantMarker, tt.shell)
			}
		})
	}
}

func TestCompletionScriptUnknownShell(t *testing.T) {
	_, err := completionScript("tcsh")
	if err == nil {
		t.Fatalf("expected unknown shell error")
	}
}
