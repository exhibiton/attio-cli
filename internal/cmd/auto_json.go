package cmd

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func shouldAutoJSON(flags RootFlags) bool {
	if flags.JSON || flags.Plain {
		return false
	}
	if !autoJSONEnabled() {
		return false
	}
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

func autoJSONEnabled() bool {
	raw := strings.TrimSpace(os.Getenv("ATTIO_AUTO_JSON"))
	if raw == "" {
		return false
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return v
}
