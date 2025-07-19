// Copyright (C) 2019-2021, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/afero"

	"github.com/luxfi/lpm/cmd"
)

func main() {
	lpm, err := cmd.New(afero.NewOsFs())
	if err != nil {
		fmt.Printf("Failed to initialize the lpm command: %s.\n", err)
		os.Exit(1)
	}

	if err := lpm.Execute(); err != nil {
		fmt.Printf("Unexpected error %s.\n", err)
		os.Exit(1)
	}
}
