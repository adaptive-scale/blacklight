package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "root command with no args",
			args:           []string{},
			expectedOutput: "Allows users to scan code, files for secrets.\nversion: 0.1.0\n",
			expectError:    false,
		},
		{
			name:           "help flag",
			args:           []string{"--help"},
			expectedOutput: "A pluggable secret scanner",
			expectError:    false,
		},
		{
			name:           "invalid flag",
			args:           []string{"--invalid-flag"},
			expectedOutput: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tt.expectedOutput)
			}
		})
	}
}

func TestToggleFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--toggle"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	err := rootCmd.Execute()
	assert.NoError(t, err)

	// Get the value of toggle flag
	toggle, err := rootCmd.Flags().GetBool("toggle")
	assert.NoError(t, err)
	assert.True(t, toggle)
} 