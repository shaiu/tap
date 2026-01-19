package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDetermineMode_HeadlessFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("headless", true, "")
	cmd.Flags().Bool("interactive", false, "")

	mode := determineMode(cmd)
	assert.Equal(t, ModeHeadless, mode)
}

func TestDetermineMode_InteractiveFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("headless", false, "")
	cmd.Flags().Bool("interactive", true, "")

	mode := determineMode(cmd)
	assert.Equal(t, ModeInteractive, mode)
}

func TestDetermineMode_HeadlessTakesPriority(t *testing.T) {
	// When both flags are set, headless should take priority
	cmd := &cobra.Command{}
	cmd.Flags().Bool("headless", true, "")
	cmd.Flags().Bool("interactive", true, "")

	mode := determineMode(cmd)
	assert.Equal(t, ModeHeadless, mode)
}

func TestDetermineMode_CIEnvVar(t *testing.T) {
	// Save and restore original value
	orig := os.Getenv("CI")
	defer os.Setenv("CI", orig)

	os.Setenv("CI", "true")

	cmd := &cobra.Command{}
	cmd.Flags().Bool("headless", false, "")
	cmd.Flags().Bool("interactive", false, "")

	mode := determineMode(cmd)
	assert.Equal(t, ModeHeadless, mode)
}

func TestDetermineMode_TapHeadlessEnvVar(t *testing.T) {
	// Save and restore original values
	origCI := os.Getenv("CI")
	origHeadless := os.Getenv("TAP_HEADLESS")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("TAP_HEADLESS", origHeadless)
	}()

	os.Unsetenv("CI")
	os.Setenv("TAP_HEADLESS", "1")

	cmd := &cobra.Command{}
	cmd.Flags().Bool("headless", false, "")
	cmd.Flags().Bool("interactive", false, "")

	mode := determineMode(cmd)
	assert.Equal(t, ModeHeadless, mode)
}

func TestDetermineMode_NoTTY(t *testing.T) {
	// When running in tests, stdin/stdout are not typically TTYs
	// So without any flags or env vars, we should get headless mode
	origCI := os.Getenv("CI")
	origHeadless := os.Getenv("TAP_HEADLESS")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("TAP_HEADLESS", origHeadless)
	}()

	os.Unsetenv("CI")
	os.Unsetenv("TAP_HEADLESS")

	cmd := &cobra.Command{}
	cmd.Flags().Bool("headless", false, "")
	cmd.Flags().Bool("interactive", false, "")

	mode := determineMode(cmd)
	// In test environment (no TTY), should return headless
	assert.Equal(t, ModeHeadless, mode)
}

func TestRootCmd_Exists(t *testing.T) {
	cmd := RootCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "tap", cmd.Use)
}

func TestRootCmd_HasRequiredFlags(t *testing.T) {
	cmd := RootCmd()

	// Check persistent flags exist
	headlessFlag := cmd.PersistentFlags().Lookup("headless")
	assert.NotNil(t, headlessFlag)
	assert.Equal(t, "bool", headlessFlag.Value.Type())

	interactiveFlag := cmd.PersistentFlags().Lookup("interactive")
	assert.NotNil(t, interactiveFlag)
	assert.Equal(t, "bool", interactiveFlag.Value.Type())

	configFlag := cmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "string", configFlag.Value.Type())

	noColorFlag := cmd.PersistentFlags().Lookup("no-color")
	assert.NotNil(t, noColorFlag)
	assert.Equal(t, "bool", noColorFlag.Value.Type())

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "bool", verboseFlag.Value.Type())
}

func TestExecutionMode_Values(t *testing.T) {
	// Ensure mode constants have expected values
	assert.Equal(t, ExecutionMode(0), ModeInteractive)
	assert.Equal(t, ExecutionMode(1), ModeHeadless)
}
