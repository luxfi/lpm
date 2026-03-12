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

func installSource(fs afero.Fs) *cobra.Command {
	var (
		vmid   string
		tag    string
		script string
		binary string
	)

	cmd := &cobra.Command{
		Use:   "install-source <owner/repo>",
		Short: "Build and install a VM plugin from source",
		Long: `Clone a GitHub repository, build from source, and install the VM plugin.

Requires Go toolchain to be installed.

Examples:
  # Build from latest source
  lpm install-source luxfi/evm

  # Build specific tag
  lpm install-source luxfi/evm --tag v0.8.35

  # Custom build script and binary path
  lpm install-source myorg/myvm --script "make build" --binary "build/myvm"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := args[0]
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid repo: %s (expected owner/repo)", repo)
			}

			pluginDir := viper.GetString(pluginPathKey)

			wf := workflow.NewInstallSource(workflow.InstallSourceConfig{
				Owner:       parts[0],
				Repo:        parts[1],
				Tag:         tag,
				VMID:        vmid,
				BuildScript: script,
				BinaryPath:  binary,
				OS:          runtime.GOOS,
				Arch:        runtime.GOARCH,
				PluginDir:   pluginDir,
				Fs:          fs,
			})

			return wf.Execute()
		},
	}

	cmd.Flags().StringVar(&vmid, "vmid", "", "VM ID (auto-detected from repo name if not set)")
	cmd.Flags().StringVar(&tag, "tag", "", "Git tag or branch to build (default: main)")
	cmd.Flags().StringVar(&script, "script", "", "Build script/command (default: auto-detect)")
	cmd.Flags().StringVar(&binary, "binary", "", "Path to built binary relative to repo root")

	return cmd
}
