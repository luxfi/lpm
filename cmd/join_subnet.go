// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func joinChain(fs afero.Fs) *cobra.Command {
	chain := ""

	command := &cobra.Command{
		Use:   "join-chain",
		Short: "Installs all virtual machines for a chain.",
	}

	command.PersistentFlags().StringVar(&chain, "chain", "", "chain alias to join")
	err := command.MarkPersistentFlagRequired("chain")
	if err != nil {
		panic(err)
	}

	command.RunE = func(_ *cobra.Command, _ []string) error {
		lpm, err := initLPM(fs)
		if err != nil {
			return err
		}

		return lpm.JoinChain(chain)
	}

	return command
}
