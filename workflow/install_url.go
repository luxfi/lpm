// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luxfi/filesystem/perms"
	"github.com/spf13/afero"

	"github.com/luxfi/lpm/checksum"
)

var _ Workflow = &InstallURL{}

// InstallURLConfig configures a direct URL binary install.
type InstallURLConfig struct {
	URL       string
	VMID      string
	SHA256    string
	PluginDir string
	Fs        afero.Fs
}

// InstallURL downloads a pre-compiled binary from a direct URL.
type InstallURL struct {
	url       string
	vmid      string
	sha256    string
	pluginDir string
	fs        afero.Fs
}

// NewInstallURL creates a new URL install workflow.
func NewInstallURL(config InstallURLConfig) *InstallURL {
	return &InstallURL{
		url:       config.URL,
		vmid:      config.VMID,
		sha256:    config.SHA256,
		pluginDir: config.PluginDir,
		fs:        config.Fs,
	}
}

// Execute runs the URL install workflow.
func (u *InstallURL) Execute() error {
	fmt.Printf("Downloading plugin from %s...\n", u.url)

	// Download to temp file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("lpm-url-%s", u.vmid))
	if err := downloadFile(u.url, tmpFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Verify checksum if provided
	if u.sha256 != "" {
		fmt.Printf("Verifying checksum...\n")
		checksummer := checksum.NewSHA256(u.fs)
		hash := fmt.Sprintf("%x", checksummer.Checksum(tmpFile))
		if hash != u.sha256 {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", u.sha256, hash)
		}
		fmt.Printf("Checksum verified: %s\n", hash)
	}

	// Ensure plugin directory exists
	if err := u.fs.MkdirAll(u.pluginDir, perms.ReadWriteExecute); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Move binary to plugin directory
	destPath := filepath.Join(u.pluginDir, u.vmid)
	if err := copyFile(tmpFile, destPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0o755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Printf("Installed plugin\n")
	fmt.Printf("  VMID:   %s\n", u.vmid)
	fmt.Printf("  Binary: %s\n", destPath)

	return nil
}
