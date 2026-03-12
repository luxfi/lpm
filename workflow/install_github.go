// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/luxfi/filesystem/perms"
	"github.com/spf13/afero"
)

var _ Workflow = &InstallGitHub{}

// InstallGitHubConfig configures a GitHub release binary install.
type InstallGitHubConfig struct {
	Owner     string
	Repo      string
	Tag       string
	VMID      string
	Pattern   string
	OS        string
	Arch      string
	PluginDir string
	Fs        afero.Fs
}

// InstallGitHub downloads a pre-compiled binary from GitHub releases.
type InstallGitHub struct {
	owner     string
	repo      string
	tag       string
	vmid      string
	pattern   string
	goos      string
	goarch    string
	pluginDir string
	fs        afero.Fs
}

// NewInstallGitHub creates a new GitHub release install workflow.
func NewInstallGitHub(config InstallGitHubConfig) *InstallGitHub {
	return &InstallGitHub{
		owner:     config.Owner,
		repo:      config.Repo,
		tag:       config.Tag,
		vmid:      config.VMID,
		pattern:   config.Pattern,
		goos:      config.OS,
		goarch:    config.Arch,
		pluginDir: config.PluginDir,
		fs:        config.Fs,
	}
}

// ghRelease is a minimal GitHub release API response.
type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

// ghAsset is a GitHub release asset.
type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Execute runs the GitHub release install workflow.
func (g *InstallGitHub) Execute() error {
	// Resolve VMID
	vmid := g.vmid
	if vmid == "" {
		id, err := resolveVMID(g.owner, g.repo)
		if err != nil {
			return fmt.Errorf("could not determine VMID: %w (use --vmid to specify)", err)
		}
		vmid = id
	}

	// Get release info
	release, err := g.getRelease()
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	fmt.Printf("Release: %s/%s %s (%d assets)\n", g.owner, g.repo, release.TagName, len(release.Assets))

	// Find the right binary for this platform
	asset, err := g.findAsset(release)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s (%d MB)...\n", asset.Name, asset.Size/(1024*1024))

	// Download to temp file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("lpm-%s-%s", g.repo, asset.Name))
	if err := downloadFile(asset.BrowserDownloadURL, tmpFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Ensure plugin directory exists
	if err := g.fs.MkdirAll(g.pluginDir, perms.ReadWriteExecute); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Move binary to plugin directory
	destPath := filepath.Join(g.pluginDir, vmid)
	if err := copyFile(tmpFile, destPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0o755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Printf("Installed %s/%s %s\n", g.owner, g.repo, release.TagName)
	fmt.Printf("  VMID:   %s\n", vmid)
	fmt.Printf("  Binary: %s\n", destPath)

	return nil
}

func (g *InstallGitHub) getRelease() (*ghRelease, error) {
	var url string
	if g.tag != "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", g.owner, g.repo, g.tag)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.owner, g.repo)
	}

	resp, err := http.Get(url) // #nosec G107 -- URL is constructed from user input for repo
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

func (g *InstallGitHub) findAsset(release *ghRelease) (*ghAsset, error) {
	// Build search patterns for this platform
	patterns := g.buildPatterns()

	// Try each pattern against assets
	for _, pattern := range patterns {
		for i := range release.Assets {
			name := strings.ToLower(release.Assets[i].Name)
			if strings.Contains(name, pattern) {
				return &release.Assets[i], nil
			}
		}
	}

	// Show available assets for debugging
	fmt.Printf("Available assets for %s:\n", release.TagName)
	for _, a := range release.Assets {
		fmt.Printf("  - %s (%d bytes)\n", a.Name, a.Size)
	}

	return nil, fmt.Errorf("no matching binary found for %s/%s (os=%s, arch=%s)", g.goos, g.goarch, g.goos, g.goarch)
}

func (g *InstallGitHub) buildPatterns() []string {
	if g.pattern != "" {
		// Use custom pattern with substitution
		p := strings.ReplaceAll(g.pattern, "{os}", g.goos)
		p = strings.ReplaceAll(p, "{arch}", g.goarch)
		return []string{strings.ToLower(p)}
	}

	// Build common patterns
	os := g.goos
	arch := g.goarch

	// Normalize arch names
	archNames := []string{arch}
	switch arch {
	case "amd64":
		archNames = append(archNames, "x86_64", "x64")
	case "arm64":
		archNames = append(archNames, "aarch64")
	}

	var patterns []string
	for _, a := range archNames {
		// e.g., "evm-plugin-linux-amd64", "myvm-linux-x86_64"
		patterns = append(patterns,
			fmt.Sprintf("%s-%s", os, a),
			fmt.Sprintf("plugin-%s-%s", os, a),
			fmt.Sprintf("%s_%s", os, a),
		)
	}

	return patterns
}

// resolveVMID tries to determine the VMID from the repo name.
func resolveVMID(owner, repo string) (string, error) {
	// Well-known VMs
	knownVMs := map[string]string{
		"luxfi/evm":       "mgj786NP7uDwBCcq6YwThhaN8FLyybkCa4zBWTQbNgmK6k9A6",
		"luxfi/coreth":    "mgj786NP7uDwBCcq6YwThhaN8FLyybkCa4zBWTQbNgmK6k9A6",
		"luxfi/subnet-evm": "srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy",
	}

	key := fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(repo))
	if id, ok := knownVMs[key]; ok {
		return id, nil
	}

	// Compute from name
	vmName := repo
	id, err := ComputeVMID(vmName)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// downloadFile downloads a URL to a local file.
func downloadFile(url, dest string) error {
	resp, err := http.Get(url) // #nosec G107 -- URL from GitHub API
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded %d bytes\n", written)
	return nil
}

// copyFile copies a file to a destination path.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}
