// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

type Executor interface {
	Execute(Workflow) error
}
