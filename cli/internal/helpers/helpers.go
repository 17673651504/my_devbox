// Package helpers provides shared command-line utilities: input validation,
// POSIX shell quoting and stdin detection, kept free of cobra concerns.
package helpers

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/huaweicloud/devspace-devbox/cli/internal/api"
	"github.com/huaweicloud/devspace-devbox/cli/internal/i18n"
)

// ValidateTimeoutSeconds validates the timeout/duration input: the valid range is 10-86400 seconds,
// an explicit 0 is invalid (not conflated with omission).
func ValidateTimeoutSeconds(v int32) error {
	if v < 10 || v > 86400 {
		return fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidTimeout), v)
	}
	return nil
}

// ValidateSandboxID validates the sandboxID input: lowercase letters, digits and hyphens, 36 characters.
func ValidateSandboxID(s string) error {
	if len(s) != 36 {
		return fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidSandboxID), s)
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidSandboxID), s)
		}
	}
	return nil
}

// ValidateTemplateID validates the templateID input: lowercase letters, digits and hyphens,
// at most 64 characters, and must not be empty.
func ValidateTemplateID(s string) error {
	if s == "" || len(s) > 64 {
		return fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidTemplateID), s)
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidTemplateID), s)
		}
	}
	return nil
}

// ParseStates parses the repeatable --state values and validates them.
func ParseStates(values []string) ([]api.SandboxState, error) {
	states := make([]api.SandboxState, 0, len(values))
	for _, s := range values {
		st := api.SandboxState(s)
		if !st.Valid() {
			return nil, fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidState), s)
		}
		states = append(states, st)
	}
	return states, nil
}

// unsafeShellChar matches the characters that need shell quoting (aligned with e2b's UNSAFE_SHELL_CHAR).
var unsafeShellChar = regexp.MustCompile(`[^\w@%+=:,./-]`)

// ShellQuote quotes a single argument per POSIX rules.
func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if unsafeShellChar.MatchString(s) {
		return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
	}
	return s
}

// PipedStdin reports whether the local stdin is a pipe/FIFO/socket/redirected file (aligned with e2b's
// isPipedStdin); an interactive terminal (character device) returns false.
func PipedStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return false
	}
	return info.Mode()&(os.ModeNamedPipe|os.ModeSocket) != 0 || info.Mode().IsRegular()
}
