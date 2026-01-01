// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/luxfi/lpm/workflow"
)

func link(fs afero.Fs) *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "link <org/name> <path>",
		Short: "Link a local VM binary for development",
		Long: `Link a local VM binary to the plugins directory for development.

Creates a symlink for a locally built VM binary in the plugins/current directory.
Use this during development to test local builds with the node.

Package format: <org>/<name> (e.g., luxfi/evm, myuser/myvm)

The binary must exist and be executable.

Examples:
  lpm link luxfi/evm ~/work/lux/evm/build/evm
  lpm link luxfi/evm ~/work/lux/evm/build/evm --version v1.2.3-dev
  lpm link myuser/myvm /path/to/myvm/build/myvm`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgRef := args[0]
			binaryPath := args[1]

			// Parse org/name
			parts := strings.SplitN(pkgRef, "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid package reference: %s (expected org/name)", pkgRef)
			}
			org, name := parts[0], parts[1]

			// Expand ~ in path
			if strings.HasPrefix(binaryPath, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to expand home directory: %w", err)
				}
				binaryPath = filepath.Join(home, binaryPath[1:])
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(binaryPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}

			// Determine version
			if version == "" {
				version = "v0.0.0-local"
			}

			// Get plugin directory
			pluginDir := viper.GetString(pluginPathKey)

			linkWorkflow := workflow.NewLink(workflow.LinkConfig{
				Org:        org,
				Name:       name,
				Version:    version,
				BinaryPath: absPath,
				PluginDir:  pluginDir,
				Fs:         fs,
			})

			return linkWorkflow.Execute()
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "Version label (default: v0.0.0-local)")

	return cmd
}
