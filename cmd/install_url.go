// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/luxfi/lpm/workflow"
)

func installURL(fs afero.Fs) *cobra.Command {
	var (
		vmid   string
		sha256 string
	)

	cmd := &cobra.Command{
		Use:   "install-url <url>",
		Short: "Install a VM plugin binary from a direct URL",
		Long: `Download and install a pre-compiled VM plugin binary from any URL.

The URL should point to a raw executable binary (not an archive).

Examples:
  # Install binary with explicit VMID
  lpm install-url https://github.com/luxfi/evm/releases/download/v0.8.35/evm-plugin-linux-amd64 \
    --vmid mgj786NP7uDwBCcq6YwThhaN8FLyybkCa4zBWTQbNgmK6k9A6

  # Install with SHA256 verification
  lpm install-url https://example.com/myvm-linux-amd64 \
    --vmid rWhpmtaWqaYAPFMUa4gJSXrGATiuGQVje51kkAFag1sQqK3Kn \
    --sha256 abc123...`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if vmid == "" {
				return fmt.Errorf("--vmid is required for direct URL installs")
			}

			pluginDir := viper.GetString(pluginPathKey)

			wf := workflow.NewInstallURL(workflow.InstallURLConfig{
				URL:       args[0],
				VMID:      vmid,
				SHA256:    sha256,
				PluginDir: pluginDir,
				Fs:        fs,
			})

			return wf.Execute()
		},
	}

	cmd.Flags().StringVar(&vmid, "vmid", "", "VM ID (required)")
	cmd.Flags().StringVar(&sha256, "sha256", "", "Expected SHA256 checksum (optional)")
	_ = cmd.MarkFlagRequired("vmid")

	return cmd
}
