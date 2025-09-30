// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package types

type Definition interface {
	GetAlias() string
	GetHomepage() string
	GetDescription() string
	GetMaintainers() []string
}
