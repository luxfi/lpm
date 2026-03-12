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

func installGitlab(fs afero.Fs) *cobra.Command {
	var (
		vmid      string
		tag       string
		pattern   string
		gitlabURL string
		token     string
	)

	cmd := &cobra.Command{
		Use:   "install-gitlab <owner/repo>",
		Short: "Install a VM plugin binary from a GitLab release",
		Long: `Download and install a pre-compiled VM plugin binary from GitLab releases.

Supports gitlab.com and self-hosted GitLab instances.
Automatically detects the correct binary for your platform (OS/arch).

Examples:
  # Install latest release from gitlab.com
  lpm install-gitlab myorg/myvm

  # Install specific version
  lpm install-gitlab myorg/myvm --tag v1.0.0

  # Self-hosted GitLab
  lpm install-gitlab myorg/myvm --gitlab-url https://gitlab.example.com

  # Private repository
  lpm install-gitlab myorg/myvm --token glpat-xxxxxxxxxxxxxxxxxxxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			repo := args[0]
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid repo: %s (expected owner/repo)", repo)
			}

			pluginDir := viper.GetString(pluginPathKey)

			wf := workflow.NewInstallGitLab(workflow.InstallGitLabConfig{
				Owner:     parts[0],
				Repo:      parts[1],
				Tag:       tag,
				VMID:      vmid,
				Pattern:   pattern,
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
				PluginDir: pluginDir,
				BaseURL:   gitlabURL,
				Token:     token,
				Fs:        fs,
			})

			return wf.Execute()
		},
	}

	cmd.Flags().StringVar(&vmid, "vmid", "", "VM ID (auto-detected from repo name if not set)")
	cmd.Flags().StringVar(&tag, "tag", "", "Release tag (default: latest)")
	cmd.Flags().StringVar(&pattern, "pattern", "", "Binary name pattern (default: auto-detect)")
	cmd.Flags().StringVar(&gitlabURL, "gitlab-url", "", "GitLab instance URL (default: https://gitlab.com)")
	cmd.Flags().StringVar(&token, "token", "", "GitLab private token for authentication")

	return cmd
}
