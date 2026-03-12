// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/luxfi/filesystem/perms"
	"github.com/spf13/afero"
)

var _ Workflow = &InstallSource{}

// InstallSourceConfig configures a build-from-source install.
type InstallSourceConfig struct {
	Owner       string
	Repo        string
	Tag         string
	VMID        string
	BuildScript string
	BinaryPath  string
	OS          string
	Arch        string
	PluginDir   string
	Fs          afero.Fs
}

// InstallSource clones, builds, and installs a VM from source.
type InstallSource struct {
	owner       string
	repo        string
	tag         string
	vmid        string
	buildScript string
	binaryPath  string
	goos        string
	goarch      string
	pluginDir   string
	fs          afero.Fs
}

// NewInstallSource creates a new source build install workflow.
func NewInstallSource(config InstallSourceConfig) *InstallSource {
	return &InstallSource{
		owner:       config.Owner,
		repo:        config.Repo,
		tag:         config.Tag,
		vmid:        config.VMID,
		buildScript: config.BuildScript,
		binaryPath:  config.BinaryPath,
		goos:        config.OS,
		goarch:      config.Arch,
		pluginDir:   config.PluginDir,
		fs:          config.Fs,
	}
}

// Execute runs the source install workflow.
func (s *InstallSource) Execute() error {
	// Resolve VMID
	vmid := s.vmid
	if vmid == "" {
		id, err := resolveVMID(s.owner, s.repo)
		if err != nil {
			return fmt.Errorf("could not determine VMID: %w (use --vmid to specify)", err)
		}
		vmid = id
	}

	// Check Go is available
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go toolchain not found: %w\nInstall Go from https://go.dev/dl/ or use 'lpm install-github' for pre-compiled binaries", err)
	}

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("lpm-src-%s-", s.repo))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone repo
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", s.owner, s.repo)
	ref := s.tag
	if ref == "" {
		ref = "main"
	}

	fmt.Printf("Cloning %s@%s...\n", cloneURL, ref)
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", ref, cloneURL, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try without --branch (maybe ref is a commit)
		fmt.Printf("Trying clone without branch specification...\n")
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0o755)
		cmd = exec.Command("git", "clone", "--depth", "1", cloneURL, tmpDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	}

	// Determine build command
	buildCmd := s.buildScript
	if buildCmd == "" {
		buildCmd = s.detectBuildCommand(tmpDir)
	}

	fmt.Printf("Building with: %s\n", buildCmd)

	// Run build
	parts := strings.Fields(buildCmd)
	build := exec.Command(parts[0], parts[1:]...) // #nosec G204 -- build scripts are trusted
	build.Dir = tmpDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	build.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", s.goos),
		fmt.Sprintf("GOARCH=%s", s.goarch),
		"CGO_ENABLED=0",
	)

	if err := build.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Find the binary
	binaryPath := s.binaryPath
	if binaryPath == "" {
		binaryPath = s.findBinary(tmpDir)
	}
	if binaryPath == "" {
		return fmt.Errorf("could not find built binary (use --binary to specify path)")
	}

	fullBinaryPath := filepath.Join(tmpDir, binaryPath)
	if _, err := os.Stat(fullBinaryPath); err != nil {
		return fmt.Errorf("built binary not found at %s: %w", fullBinaryPath, err)
	}

	// Install
	if err := s.fs.MkdirAll(s.pluginDir, perms.ReadWriteExecute); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	destPath := filepath.Join(s.pluginDir, vmid)
	if err := copyFile(fullBinaryPath, destPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	if err := os.Chmod(destPath, 0o755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Printf("Built and installed %s/%s\n", s.owner, s.repo)
	fmt.Printf("  VMID:   %s\n", vmid)
	fmt.Printf("  Binary: %s\n", destPath)

	return nil
}

func (s *InstallSource) detectBuildCommand(dir string) string {
	// Check for Makefile
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
		return "make build"
	}

	// Check for scripts/build.sh
	if _, err := os.Stat(filepath.Join(dir, "scripts", "build.sh")); err == nil {
		return "./scripts/build.sh"
	}

	// Default Go build
	if _, err := os.Stat(filepath.Join(dir, "plugin")); err == nil {
		return "go build -o build/plugin ./plugin"
	}

	return "go build -o build/plugin ./..."
}

func (s *InstallSource) findBinary(dir string) string {
	// Check common locations
	candidates := []string{
		"build/plugin",
		"build/" + s.repo,
		"plugin/" + s.repo,
		s.repo,
	}

	for _, c := range candidates {
		path := filepath.Join(dir, c)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return c
		}
	}

	return ""
}
