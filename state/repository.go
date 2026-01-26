// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/luxfi/lpm/git"
	"github.com/luxfi/lpm/types"
)

var (
	vmDir    = "vms"
	chainDir = "chains"

	extension = "yaml"
)

// Repository wraps a plugin repository's VMs and Chains
type Repository interface {
	GetPath() string
	GetVM(name string) (Definition[types.VM], error)
	GetChain(name string) (Definition[types.Chain], error)
}

type DiskRepository struct {
	Git  git.Factory
	Path string
}

func (d DiskRepository) GetVM(name string) (Definition[types.VM], error) {
	return get[types.VM](d, vmDir, name)
}

func (d DiskRepository) GetChain(name string) (Definition[types.Chain], error) {
	return get[types.Chain](d, chainDir, name)
}

func (d DiskRepository) GetPath() string {
	return d.Path
}

func get[T types.Definition](d DiskRepository, dir string, file string) (Definition[T], error) {
	relativePathWithExtension := filepath.Join(dir, fmt.Sprintf("%s.%s", file, extension))
	absolutePathWithExtension := filepath.Join(d.Path, relativePathWithExtension)
	bytes, err := os.ReadFile(absolutePathWithExtension)
	if err != nil {
		return Definition[T]{}, err
	}

	var definition T
	if err := yaml.Unmarshal(bytes, &definition); err != nil {
		return Definition[T]{}, err
	}

	commit, err := d.Git.GetLastModified(d.Path, relativePathWithExtension)
	if err != nil {
		return Definition[T]{}, err
	}

	return Definition[T]{
		Definition: definition,
		Commit:     commit,
	}, nil
}
