// Copyright (C) 2019-2021, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package engine

import (
	"fmt"

	"github.com/luxdefi/apm/state"
	"github.com/luxdefi/apm/workflow"
)

var _ workflow.Executor = &WorkflowEngine{}

func NewWorkflowEngine(stateFile state.File) *WorkflowEngine {
	return &WorkflowEngine{
		stateFile: stateFile,
	}
}

type WorkflowEngine struct {
	stateFile state.File
}

func (w *WorkflowEngine) Execute(workflow workflow.Workflow) error {
	defer func() {
		if err := w.stateFile.Commit(); err != nil {
			fmt.Printf("failed to commit the statefile")
		}
	}()

	return workflow.Execute()
}
