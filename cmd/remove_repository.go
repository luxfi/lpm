// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func removeRepository(fs afero.Fs) *cobra.Command {
	alias := ""

	command := &cobra.Command{
		Use:   "remove-repository",
		Short: "removes a repository from the list of tracked repositories",
	}
	command.PersistentFlags().StringVar(&alias, "alias", "", "alias for the repository")
	err := command.MarkPersistentFlagRequired("alias")
	if err != nil {
		// TODO cleanup these panics
		panic(err)
	}

	command.RunE = func(_ *cobra.Command, _ []string) error {
		lpm, err := initLPM(fs)
		if err != nil {
			return err
		}

		return lpm.RemoveRepository(alias)
	}

	return command
}
