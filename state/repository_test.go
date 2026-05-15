// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/lpm/types"
)

// TestDecodeWrappedVM exercises the canonical on-disk format used by every
// shipped registry entry (top-level `vm:` wrapper). Regression test for the
// silent zero-value unmarshal documented in CI-PREMORTEM.md §12a.
func TestDecodeWrappedVM(t *testing.T) {
	yamlBytes := []byte(`vm:
  id: "mgj786NP7uDwBCcq6YwThhaN8FLyybkCa4zBWTQbNgmK6k9A6"
  alias: "cevm-darwin-arm64"
  homepage: "https://github.com/luxfi/luxcpp"
  description: "cevm test"
  maintainers:
    - "z@lux.network"
  installScript: "scripts/install.sh"
  binaryPath: "cevm"
  url: "https://example.com/cevm.tar.gz"
  sha256: "0000000000000000000000000000000000000000000000000000000000000000"
  version:
    major: 0
    minor: 1
    patch: 0
`)
	vm, err := decodeWrapped[types.VM](yamlBytes, vmWrapperKey)
	require.NoError(t, err)
	require.Equal(t, "mgj786NP7uDwBCcq6YwThhaN8FLyybkCa4zBWTQbNgmK6k9A6", vm.ID)
	require.Equal(t, "cevm-darwin-arm64", vm.Alias)
	require.Equal(t, "https://example.com/cevm.tar.gz", vm.URL)
	require.Equal(t, "scripts/install.sh", vm.InstallScript)
	require.Equal(t, "cevm", vm.BinaryPath)
	require.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", vm.SHA256)
}

func TestDecodeWrappedVMMissingKey(t *testing.T) {
	yamlBytes := []byte(`id: "x"
alias: "y"
url: "z"
`)
	_, err := decodeWrapped[types.VM](yamlBytes, vmWrapperKey)
	require.Error(t, err, "must reject unwrapped manifests instead of returning zero VM")
}

func TestDecodeWrappedChain(t *testing.T) {
	yamlBytes := []byte(`subnet:
  id:
    mainnet: "abc"
  alias: "wagmi"
  homepage: "https://example.com"
  description: "test"
  maintainers:
    - "z@lux.network"
  vms:
    - "evm"
`)
	chain, err := decodeWrapped[types.Chain](yamlBytes, chainWrapperKey)
	require.NoError(t, err)
	require.Equal(t, "wagmi", chain.Alias)
	require.Equal(t, []string{"evm"}, chain.VMs)
	require.Equal(t, "abc", chain.ID["mainnet"])
}

func TestDecodeWrappedRejectsEmpty(t *testing.T) {
	_, err := decodeWrapped[types.VM]([]byte(``), vmWrapperKey)
	require.Error(t, err)
}
