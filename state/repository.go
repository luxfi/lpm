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

	// Wrapper keys used by on-disk plugin manifests. By convention every
	// definition file is a single-key mapping (`vm:` for VMs, `subnet:` for
	// chains/subnets) whose value is the actual definition. The keys match
	// the wrappers in plugins-core/examples/{vm,subnet}.yaml and every
	// shipped registry entry.
	vmWrapperKey    = "vm"
	chainWrapperKey = "subnet"
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
	return get[types.VM](d, vmDir, name, vmWrapperKey)
}

func (d DiskRepository) GetChain(name string) (Definition[types.Chain], error) {
	return get[types.Chain](d, chainDir, name, chainWrapperKey)
}

func (d DiskRepository) GetPath() string {
	return d.Path
}

func get[T types.Definition](d DiskRepository, dir string, file string, wrapperKey string) (Definition[T], error) {
	relativePathWithExtension := filepath.Join(dir, fmt.Sprintf("%s.%s", file, extension))
	absolutePathWithExtension := filepath.Join(d.Path, relativePathWithExtension)
	bytes, err := os.ReadFile(absolutePathWithExtension)
	if err != nil {
		return Definition[T]{}, err
	}

	definition, err := decodeWrapped[T](bytes, wrapperKey)
	if err != nil {
		return Definition[T]{}, fmt.Errorf("decoding %s: %w", relativePathWithExtension, err)
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

// decodeWrapped unmarshals a definition YAML that wraps its payload under a
// single top-level key (e.g. `vm:` or `subnet:`). Empty or missing wrappers
// are reported as errors so callers do not silently get a zero-valued T and
// crash later (see CI-PREMORTEM.md §12a).
func decodeWrapped[T types.Definition](data []byte, wrapperKey string) (T, error) {
	var zero T
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return zero, err
	}
	// A document node has exactly one content child: the top-level mapping.
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 {
		return zero, fmt.Errorf("expected a single YAML document, got kind=%d content=%d", root.Kind, len(root.Content))
	}
	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return zero, fmt.Errorf("expected top-level mapping, got kind=%d", mapping.Kind)
	}
	// MappingNode.Content is a flat [key0, value0, key1, value1, ...] list.
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i]
		if key.Kind == yaml.ScalarNode && key.Value == wrapperKey {
			var out T
			if err := mapping.Content[i+1].Decode(&out); err != nil {
				return zero, err
			}
			return out, nil
		}
	}
	return zero, fmt.Errorf("missing required top-level %q key", wrapperKey)
}
