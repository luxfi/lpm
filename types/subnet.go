// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package types

var _ Definition = &Chain{}

type Chain struct {
	ID          map[string]string `yaml:"id"`
	Alias       string            `yaml:"alias"`
	Homepage    string            `yaml:"homepage"`
	Description string            `yaml:"description"`
	Maintainers []string          `yaml:"maintainers"`
	VMs         []string          `yaml:"vms"`
	// Config      chains.ChainConfig `yaml:"config,omitempty"`
}

func (s Chain) GetID(network string) (string, bool) {
	id, ok := s.ID[network]
	return id, ok
}

func (s Chain) GetAlias() string {
	return s.Alias
}

func (s Chain) GetHomepage() string {
	return s.Homepage
}

func (s Chain) GetDescription() string {
	return s.Description
}

func (s Chain) GetMaintainers() []string {
	return s.Maintainers
}
