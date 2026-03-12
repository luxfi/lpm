// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/luxfi/lpm/workflow"
)

func installGithub(fs afero.Fs) *cobra.Command {
	var (
		vmid    string
		tag     string
		pattern string
	)

	cmd := &cobra.Command{
		Use:   "install-github <owner/repo>",
		Short: "Install a VM plugin binary from a GitHub release",
		Long: `Download and install a pre-compiled VM plugin binary from GitHub releases.

Automatically detects the correct binary for your platform (OS/arch).

Examples:
  # Install latest release, auto-detect binary
  lpm install-github luxfi/evm

  # Install specific version
  lpm install-github luxfi/evm --tag v0.8.35

  # Install with explicit VMID
  lpm install-github luxfi/evm --vmid mgj786NP7uDwBCcq6YwThhaN8FLyybkCa4zBWTQbNgmK6k9A6

  # Custom binary name pattern
  lpm install-github myorg/myvm --pattern "myvm-plugin-{os}-{arch}"`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			repo := args[0]
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid repo: %s (expected owner/repo)", repo)
			}

			pluginDir := viper.GetString(pluginPathKey)

			wf := workflow.NewInstallGitHub(workflow.InstallGitHubConfig{
				Owner:     parts[0],
				Repo:      parts[1],
				Tag:       tag,
				VMID:      vmid,
				Pattern:   pattern,
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
				PluginDir: pluginDir,
				Fs:        fs,
			})

			return wf.Execute()
		},
	}

	cmd.Flags().StringVar(&vmid, "vmid", "", "VM ID (auto-detected from repo name if not set)")
	cmd.Flags().StringVar(&tag, "tag", "", "Release tag (default: latest)")
	cmd.Flags().StringVar(&pattern, "pattern", "", "Binary name pattern (default: auto-detect)")

	return cmd
}
