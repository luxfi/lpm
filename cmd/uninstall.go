// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func uninstall(fs afero.Fs) *cobra.Command {
	vm := ""
	command := &cobra.Command{
		Use:   "uninstall-vm",
		Short: "Uninstalls a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vm, "vm", "", "vm alias to install")
	err := command.MarkPersistentFlagRequired("vm")
	if err != nil {
		panic(err)
	}

	command.RunE = func(_ *cobra.Command, _ []string) error {
		lpm, err := initLPM(fs)
		if err != nil {
			return err
		}

		return lpm.Uninstall(vm)
	}

	return command
}
