// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/luxfi/filesystem/perms"
	"github.com/spf13/afero"
)

var _ Workflow = &InstallGitLab{}

// InstallGitLabConfig configures a GitLab release binary install.
type InstallGitLabConfig struct {
	Owner     string
	Repo      string
	Tag       string
	VMID      string
	Pattern   string
	OS        string
	Arch      string
	PluginDir string
	BaseURL   string // GitLab instance URL (default: https://gitlab.com)
	Token     string // Private token for authentication
	Fs        afero.Fs
}

// InstallGitLab downloads a pre-compiled binary from GitLab releases.
type InstallGitLab struct {
	owner     string
	repo      string
	tag       string
	vmid      string
	pattern   string
	goos      string
	goarch    string
	pluginDir string
	baseURL   string
	token     string
	fs        afero.Fs
}

// NewInstallGitLab creates a new GitLab release install workflow.
func NewInstallGitLab(config InstallGitLabConfig) *InstallGitLab {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	return &InstallGitLab{
		owner:     config.Owner,
		repo:      config.Repo,
		tag:       config.Tag,
		vmid:      config.VMID,
		pattern:   config.Pattern,
		goos:      config.OS,
		goarch:    config.Arch,
		pluginDir: config.PluginDir,
		baseURL:   strings.TrimRight(baseURL, "/"),
		token:     config.Token,
		fs:        config.Fs,
	}
}

// glRelease is a minimal GitLab release API response.
type glRelease struct {
	TagName string   `json:"tag_name"`
	Assets  glAssets `json:"assets"`
}

// glAssets contains GitLab release assets.
type glAssets struct {
	Links   []glLink   `json:"links"`
	Sources []glSource `json:"sources"`
}

// glLink is a GitLab release link (uploaded binary).
type glLink struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	LinkType string `json:"link_type"` // "other", "runbook", "image", "package"
}

// glSource is a GitLab release source archive.
type glSource struct {
	Format string `json:"format"`
	URL    string `json:"url"`
}

// Execute runs the GitLab release install workflow.
func (g *InstallGitLab) Execute() error {
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

	fmt.Printf("Release: %s/%s %s (%d assets)\n", g.owner, g.repo, release.TagName, len(release.Assets.Links))

	// Find the right binary for this platform
	link, err := g.findAsset(release)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", link.Name)

	// Download to temp file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("lpm-%s-%s", g.repo, link.Name))
	if err := g.downloadWithAuth(link.URL, tmpFile); err != nil {
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

func (g *InstallGitLab) getRelease() (*glRelease, error) {
	// GitLab requires URL-encoded project path
	projectPath := url.PathEscape(fmt.Sprintf("%s/%s", g.owner, g.repo))

	var apiURL string
	if g.tag != "" {
		apiURL = fmt.Sprintf("%s/api/v4/projects/%s/releases/%s", g.baseURL, projectPath, g.tag)
	} else {
		// GitLab returns releases sorted by created_at desc, first is latest
		apiURL = fmt.Sprintf("%s/api/v4/projects/%s/releases", g.baseURL, projectPath)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if g.token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API returned %d: %s", resp.StatusCode, string(body))
	}

	if g.tag != "" {
		var release glRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf("failed to parse release: %w", err)
		}
		return &release, nil
	}

	// For latest, decode array and take the first
	var releases []glRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found for %s/%s", g.owner, g.repo)
	}
	return &releases[0], nil
}

func (g *InstallGitLab) findAsset(release *glRelease) (*glLink, error) {
	patterns := g.buildPatterns()

	for _, pattern := range patterns {
		for i := range release.Assets.Links {
			name := strings.ToLower(release.Assets.Links[i].Name)
			if strings.Contains(name, pattern) {
				return &release.Assets.Links[i], nil
			}
		}
	}

	// Show available assets for debugging
	fmt.Printf("Available assets for %s:\n", release.TagName)
	for _, l := range release.Assets.Links {
		fmt.Printf("  - %s (%s)\n", l.Name, l.URL)
	}

	return nil, fmt.Errorf("no matching binary found for %s/%s (os=%s, arch=%s)", g.goos, g.goarch, g.goos, g.goarch)
}

func (g *InstallGitLab) buildPatterns() []string {
	if g.pattern != "" {
		p := strings.ReplaceAll(g.pattern, "{os}", g.goos)
		p = strings.ReplaceAll(p, "{arch}", g.goarch)
		return []string{strings.ToLower(p)}
	}

	os := g.goos
	arch := g.goarch

	archNames := []string{arch}
	switch arch {
	case "amd64":
		archNames = append(archNames, "x86_64", "x64")
	case "arm64":
		archNames = append(archNames, "aarch64")
	}

	patterns := make([]string, 0, len(archNames)*3)
	for _, a := range archNames {
		patterns = append(patterns,
			fmt.Sprintf("%s-%s", os, a),
			fmt.Sprintf("plugin-%s-%s", os, a),
			fmt.Sprintf("%s_%s", os, a),
		)
	}

	return patterns
}

func (g *InstallGitLab) downloadWithAuth(dlURL, dest string) error {
	req, err := http.NewRequest("GET", dlURL, nil)
	if err != nil {
		return err
	}

	if g.token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.token)
	}

	resp, err := http.DefaultClient.Do(req)
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
