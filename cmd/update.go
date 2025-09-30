// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func update(fs afero.Fs) *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "Updates plugin definitions for all tracked repositories.",
	}
	command.RunE = func(_ *cobra.Command, _ []string) error {
		lpm, err := initLPM(fs)
		if err != nil {
			return err
		}

		return lpm.Update()
	}

	return command
}
