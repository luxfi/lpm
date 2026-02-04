// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luxfi/filesystem/perms"
	"github.com/luxfi/ids"
	"github.com/spf13/afero"
)

var _ Workflow = &Link{}

// LinkConfig contains configuration for linking a local VM binary
type LinkConfig struct {
	Org        string
	Name       string
	Version    string
	BinaryPath string
	PluginDir  string
	Fs         afero.Fs
}

// Link creates a development symlink for a local VM binary
type Link struct {
	org        string
	name       string
	version    string
	binaryPath string
	pluginDir  string
	fs         afero.Fs
}

// NewLink creates a new Link workflow
func NewLink(config LinkConfig) *Link {
	return &Link{
		org:        config.Org,
		name:       config.Name,
		version:    config.Version,
		binaryPath: config.BinaryPath,
		pluginDir:  config.PluginDir,
		fs:         config.Fs,
	}
}

// Execute runs the link workflow
func (l *Link) Execute() error {
	// Validate binary exists
	info, err := os.Stat(l.binaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("binary not found: %s", l.binaryPath)
		}
		return fmt.Errorf("failed to stat binary: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", l.binaryPath)
	}

	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("binary is not executable: %s", l.binaryPath)
	}

	// Determine VM name for VMID calculation
	vmName := l.name
	if l.name == "evm" && l.org == "luxfi" {
		vmName = "Lux EVM" // Canonical name for Lux EVM
	}

	// Calculate VMID
	vmID, err := ComputeVMID(vmName)
	if err != nil {
		return fmt.Errorf("failed to calculate VMID: %w", err)
	}

	// Ensure plugins/current directory exists
	currentDir := filepath.Join(l.pluginDir, "current")
	if err := l.fs.MkdirAll(currentDir, perms.ReadWriteExecute); err != nil {
		return fmt.Errorf("failed to create plugins/current directory: %w", err)
	}

	// Create VMID symlink in plugins/current (node compatibility)
	vmidPath := filepath.Join(currentDir, vmID.String())

	// Remove existing symlink if present
	if _, err := l.fs.Stat(vmidPath); err == nil {
		if err := l.fs.Remove(vmidPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Create symlink using OS (afero doesn't support symlinks well)
	if err := os.Symlink(l.binaryPath, vmidPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("Plugin linked successfully:\n")
	fmt.Printf("  Package:  %s/%s@%s\n", l.org, l.name, l.version)
	fmt.Printf("  VM Name:  %s\n", vmName)
	fmt.Printf("  VMID:     %s\n", vmID.String())
	fmt.Printf("  Binary:   %s\n", l.binaryPath)
	fmt.Printf("  Symlink:  %s\n", vmidPath)

	return nil
}

// ComputeVMID computes the VMID for a VM name
// VMID = CB58(pad32(vmName))
func ComputeVMID(vmName string) (ids.ID, error) {
	if len(vmName) > 32 {
		return ids.Empty, fmt.Errorf("VM name must be <= 32 bytes, found %d", len(vmName))
	}

	// Pad to 32 bytes
	b := make([]byte, 32)
	copy(b, []byte(vmName))

	// Convert to ID (CB58 encoded)
	return ids.ToID(b)
}
