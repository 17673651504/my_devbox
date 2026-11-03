package cmd

import (
	"fmt"

	"github.com/huaweicloud/devspace-devbox/cli/internal/api"
	"github.com/huaweicloud/devspace-devbox/cli/internal/i18n"
)

// validateTimeoutSeconds validates the timeout/duration input: the valid range is 10-86400 seconds,
// an explicit 0 is invalid (not conflated with omission).
func validateTimeoutSeconds(v int32) error {
	if v < 10 || v > 86400 {
		return fmt.Errorf(i18n.T(i18n.Msg.Sandbox.InvalidTimeout), v)
	}
	return nil
}

// validateSandboxID validates the sandboxID input: lowercase letters, digits and hyphens, 36 characters.
func validateSandboxID(s string) error {
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

// validateTemplateID validates the templateID input: lowercase letters, digits and hyphens,
// at most 64 characters, and must not be empty.
func validateTemplateID(s string) error {
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

// parseStates parses the repeatable --state values and validates them.
func parseStates(values []string) ([]api.SandboxState, error) {
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
